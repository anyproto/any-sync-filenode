package s3store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/anyproto/any-sync-filenode/store"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/commonfile/fileblockstore"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

const CName = fileblockstore.CName

var log = logger.NewNamed(CName)

func New() S3Store {
	return new(s3store)
}

type S3Store interface {
	store.Store
	app.ComponentRunnable
}

type s3store struct {
	bucket  *string
	client  *s3.S3
	limiter chan struct{}
	sess    *session.Session
}

func (s *s3store) Init(a *app.App) (err error) {
	conf := a.MustComponent("config").(configSource).GetS3Store()
	if conf.Profile == "" {
		conf.Profile = "default"
	}
	if conf.Bucket == "" {
		return fmt.Errorf("s3 bucket is empty")
	}
	if conf.MaxThreads <= 0 {
		conf.MaxThreads = 16
	}
	var endpoint *string
	if conf.Endpoint != "" {
		endpoint = aws.String(conf.Endpoint)
	}
	s.sess, err = session.NewSessionWithOptions(session.Options{
		Profile: conf.Profile,
		Config: aws.Config{
			Region:   aws.String(conf.Region),
			Endpoint: endpoint,
		},
	})
	s.bucket = aws.String(conf.Bucket)
	s.client = s3.New(s.sess)
	s.limiter = make(chan struct{}, conf.MaxThreads)
	return nil
}

func (s *s3store) Name() (name string) {
	return CName
}

func (s *s3store) Run(ctx context.Context) (err error) {
	return nil
}

func (s *s3store) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	st := time.Now()
	s.limiter <- struct{}{}
	defer func() { <-s.limiter }()
	wait := time.Since(st)
	obj, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(k.String()),
	})
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()
	data, err := io.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}
	log.Debug("s3 get",
		zap.Duration("total", time.Since(st)),
		zap.Duration("wait", wait),
		zap.Int("kbytes", len(data)/1024),
	)
	return blocks.NewBlockWithCid(data, k)
}

func (s *s3store) GetMany(ctx context.Context, ks []cid.Cid) <-chan blocks.Block {
	var res = make(chan blocks.Block)
	go func() {
		defer close(res)
		var wg sync.WaitGroup
		var getManyLimiter = make(chan struct{}, 4)
		for _, k := range ks {
			wg.Add(1)
			select {
			case getManyLimiter <- struct{}{}:
			case <-ctx.Done():
				return
			}
			go func(k cid.Cid) {
				defer func() { <-getManyLimiter }()
				defer wg.Done()
				b, e := s.Get(ctx, k)
				if e == nil {
					select {
					case res <- b:
					case <-ctx.Done():
					}
				} else {
					log.Info("get error", zap.Error(e))
				}
			}(k)
		}
		wg.Wait()
	}()
	return res
}

func (s *s3store) Add(ctx context.Context, bs []blocks.Block) error {
	st := time.Now()
	s.limiter <- struct{}{}
	defer func() { <-s.limiter }()
	wait := time.Since(st)
	var dataLen int
	for _, b := range bs {
		data := b.RawData()
		dataLen += len(data)
		if _, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Key:    aws.String(b.Cid().String()),
			Body:   bytes.NewReader(data),
			Bucket: s.bucket,
		}); err != nil {
			return err
		}
	}
	log.Debug("s3 put",
		zap.Duration("total", time.Since(st)),
		zap.Duration("wait", wait),
		zap.Int("blocks", len(bs)),
		zap.Int("kbytes", dataLen/1024),
	)
	return nil
}

func (s *s3store) DeleteMany(ctx context.Context, ks []cid.Cid) error {
	for _, k := range ks {
		if e := s.Delete(ctx, k); e != nil {
			log.Warn("can't delete cid", zap.Error(e))
		}
	}
	return nil
}

func (s *s3store) Delete(ctx context.Context, c cid.Cid) error {
	// TODO: make batch delete
	st := time.Now()
	s.limiter <- struct{}{}
	wait := time.Since(st)
	defer func() { <-s.limiter }()
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(c.String()),
	})
	log.Debug("s3 delete",
		zap.Duration("total", time.Since(st)),
		zap.Duration("wait", wait),
	)
	return err
}

func (s *s3store) Close(ctx context.Context) (err error) {
	return nil
}
