package main

import (
	"context"
	"github.com/anytypeio/any-sync-filenode/config"
	"github.com/anytypeio/any-sync-filenode/s3store"
	commonaccount "github.com/anytypeio/any-sync/accountservice"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/metric"
	"github.com/anytypeio/any-sync/net"
	"github.com/stretchr/testify/require"
	"testing"
)

var ctx = context.Background()

func TestBootstrap(t *testing.T) {
	c := &config.Config{
		Account: commonaccount.Config{
			PeerId:        "12D3KooWQxiZ5a7vcy4DTJa8Gy1eVUmwb5ojN4SrJC9Rjxzigw6C",
			PeerKey:       "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
			SigningKey:    "X7YT92hRIQf42tNiIlOS01p1uJiPSqcdkE6LViS8PxnhAv/28YzTctqPjm11AC63Dq7ybmwPQiBu1pwa3AxHzQ==",
			EncryptionKey: "MIIEpAIBAAKCAQEA5KuB40UIN6LbGaFw1EK7PZ30bkQerMC24hiorv9oUqBPsFXnGuqSd09B9yLWsUGBHn4q2G3ZF2e73+K6Exn/OPp4othDwUPzm0sx75y1D2iy7nNbP4EAqC4n5x3rBaLCvdN9MJEc8TfvkKJAJ8gCAub3RXuf8so5GfYra2WcFjRovjqtpV5dN0onq5Z9gNYMUrzmHNE2PjO4jpAyTh2/zIP3odpt9e14xUjB+Xv3gonX7Scl7njJFXcRMfdriDBkTDlCMPt/hA7Sabif7jXzd3hDIFBiX51SYCSeBpv70+iSNwR1bO4zKagAyI1s5MiBriNdO07CvbI6WK6MzC3NeQIDAQABAoIBAQDS9Xrt3ajYExGJEsxRtoKhNNDkzUlzXJMcAV3VnGF1INqDtqxvw4p+MYuM4QIqI2FobUM/yg+mrRfBU50QtEImIcUbjuLrMLJUSUn3YZ4UaiXxIFFFQ9EEVxiO+qXw3BhHIg5zuNx3mYAU8eq4CKf6X3QuEQAd7/w//EBQYzxdqhsGqRdawnqlRg0GdqsYP6SaGliyQcjIybb8+UA4YbBofECQmubY+ObrAKiE2Y+AxXtO/JMC4hM7r4QV9RYd9O7n5n6cP7eEN0oo6gViHEhCsMktPeawXwXXbvrS4iz/gyYFs4pYTuQxC+UNVusIRNfrHQTQuyXIqioArPgj4TlRAoGBAOUv6ioGsDF89VQFhYPrJdeoGkARRG6fRWVxuVK1Yx/3uCuHp01mjmczowY36w3MdMSZMuiRk3zeVkts3UQwHgUR8jHJT7Ty/cL80K2/n12aI2OK8vnY71idiZPoy7PL0vY2S4klLsBfumrEB9sHCE7G3+0Ai9jfePAUNj+RCHOVAoGBAP9sGiorjyZbFOCaXDemJcI140VUqJ+OFeBsiZZCOvHlgE1tFcy2RpDvs7dwe2gBb/Gkt2hsDbQdEWdSQk5AVcJBnALoWn4eoarEN6TVDxIDhN0Vz63uTYM+rMvgETVFfc8lxvEmcO7G9rxNPSkIUSqVXw3AywYizNAgRb4fRnlVAoGAY2zQ6iByqVVrXGL01BDcHt1nXenfxRnFUkfuvMnB0el2dTPpSXO3TWAiVh1GFHthILTRWAFneWE/EIOOzfkN3Oc3KZAKyxYrLj7dDLM3oLSwq4to8yjAVLIrMAZq8Hn48CUHydxelsgwqAaY3dUELqCqHjgBczknTweFrTfu8a0CgYEAtPSDPOkLS6MvkUgKmSpOid7fmpi1tgRVn1+Fwjw9wm5TjYcA8L0aFUiczBMWesK56jpF7ebPdpE5aTev3fxaRXtx6eVvZvaQlojY2yBOwvZXRMJVFeZEZ/0ZMn8V8eW/kegzn1kanS+8Uf9umwlXZ5HXe8jgjQJOlAypHF7P8NkCgYA0Z1F1kGUSl8DBIaNbOAR9FOKE/cMOubxoFCi/eHQc0wu+jEa53SnD5a8TYpEU3eqAiWsImuqoJA4RsvLZcKKfk2PJl8w+CD65YOJckyNRzUjzSz2S7oUXkUVhzbJ676y2B4Vp+CxYLllg/s5qh4ZMSOBj2quy1NhFbOq5rISddg==",
		},
		GrpcServer: net.Config{},
		Metric:     metric.Config{},
		S3Store: s3store.Config{
			Bucket: "test",
		},
		FileDevStore: config.FileDevStore{},
	}
	a := new(app.App)
	a.Register(c)
	Bootstrap(a)
	require.NoError(t, a.Start(ctx))
	require.NoError(t, a.Close(ctx))
}