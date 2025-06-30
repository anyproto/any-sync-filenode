package stat

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonfile/fileproto"
	"github.com/anyproto/any-sync/coordinator/coordinatorclient"
	"github.com/anyproto/any-sync/coordinator/coordinatorproto"

	"github.com/anyproto/any-sync-filenode/index"
)

const CName = "filenode.stat"

type accountInfoProvider interface {
	AccountInfo(ctx context.Context, identity string) (*fileproto.AccountInfoResponse, error)
	BatchAccountInfo(ctx context.Context, identities []string) ([]*fileproto.AccountInfoResponse, error)
}

type Stat interface {
	app.ComponentRunnable
}

func New() Stat {
	return &statService{}
}

type statService struct {
	accountInfoProvider accountInfoProvider
	index               index.Index
	coordinator         coordinatorclient.CoordinatorClient
}

func (i *statService) Init(a *app.App) (err error) {
	i.accountInfoProvider = app.MustComponent[accountInfoProvider](a)
	i.index = app.MustComponent[index.Index](a)
	i.coordinator = app.MustComponent[coordinatorclient.CoordinatorClient](a)
	return
}

func (i *statService) Name() (name string) {
	return CName
}

func (i *statService) Run(ctx context.Context) (err error) {
	http.HandleFunc("/stat/identity/{identity}", func(writer http.ResponseWriter, request *http.Request) {
		identity := request.PathValue("identity")
		if identity == "" {
			http.Error(writer, "identity is empty", http.StatusBadRequest)
			return
		}
		accountInfo, err := i.accountInfoProvider.AccountInfo(request.Context(), identity)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if accountInfo == nil {
			http.Error(writer, "not found", http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(accountInfo)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/stat/identities", func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			Ids []string `json:"ids"`
		}{}
		if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
			http.Error(writer, "invalid JSON", http.StatusBadRequest)
			return
		}
		accountInfos, err := i.accountInfoProvider.BatchAccountInfo(request.Context(), data.Ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(accountInfos)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/stat/check/{identity}", func(writer http.ResponseWriter, request *http.Request) {
		identity := request.PathValue("identity")
		if identity == "" {
			http.Error(writer, "identity is empty", http.StatusBadRequest)
			return
		}
		isDoFix := request.URL.Query().Get("fix") != ""

		st := time.Now()
		res, err := i.index.Check(request.Context(), index.Key{GroupId: identity}, isDoFix)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := struct {
			Results  []index.CheckResult `json:"results"`
			Duration string              `json:"duration"`
		}{
			Results:  res,
			Duration: time.Since(st).String(),
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(resp)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/stat/check_deletion/{identity}", func(writer http.ResponseWriter, request *http.Request) {
		identity := request.PathValue("identity")
		if identity == "" {
			http.Error(writer, "identity is empty", http.StatusBadRequest)
			return
		}
		isDoFix := request.URL.Query().Get("fix") != ""

		st := time.Now()
		res, err := i.index.CheckDeletedSpaces(ctx, index.Key{GroupId: identity}, func(spaceIds []string) (deletedIds []string, err error) {
			statuses, _, err := i.coordinator.StatusCheckMany(ctx, spaceIds)
			if err != nil {
				return nil, err
			}
			for idx, status := range statuses {
				if status.Status == coordinatorproto.SpaceStatus_SpaceStatusDeleted {
					deletedIds = append(deletedIds, spaceIds[idx])
				}
			}
			return
		}, isDoFix)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		resp := struct {
			Results  []string `json:"results"`
			Duration string   `json:"duration"`
		}{
			Results:  res,
			Duration: time.Since(st).String(),
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(resp)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	return nil
}

func (i *statService) Close(ctx context.Context) (err error) {
	return
}
