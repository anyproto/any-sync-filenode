package stat

import (
	"context"
	"encoding/json"
	"github.com/anyproto/any-sync/app"
	"net/http"
)

const CName = "stat.identity"

type accountInfoProvider interface {
	AccountInfoToJSON(ctx context.Context, identity string) (string, error)
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
		accountInfo, err := i.accountInfoProvider.AccountInfoToJSON(request.Context(), identity)
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
	return nil
}

func (i *identityStat) Close(ctx context.Context) (err error) {
	return
}
