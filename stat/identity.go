package stat

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonfile/fileproto"
)

const CName = "stat.identity"

type accountInfoProvider interface {
	AccountInfo(ctx context.Context, identity string) (*fileproto.AccountInfoResponse, error)
	BatchAccountInfo(ctx context.Context, identities []string) ([]*fileproto.AccountInfoResponse, error)
}

type Stat interface {
	app.ComponentRunnable
}

func New() Stat {
	return &identityStat{}
}

type identityStat struct {
	accountInfoProvider accountInfoProvider
}

func (i *identityStat) Init(a *app.App) (err error) {
	i.accountInfoProvider = app.MustComponent[accountInfoProvider](a)
	return
}

func (i *identityStat) Name() (name string) {
	return CName
}

func (i *identityStat) Run(ctx context.Context) (err error) {
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
	return nil
}

func (i *identityStat) Close(ctx context.Context) (err error) {
	return
}
