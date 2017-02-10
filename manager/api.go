package manager

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"github.com/unrolled/render"
	"net/http"
)

var (
	r = render.New(render.Options{
		IndentJSON: true,
	})
)

func Handler(man types.VolumeManager) http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/v1/volumes").Subrouter()

	s.Methods("POST").Path("/").HandlerFunc(Volume2VolumeHandlerFunc(man.Create))
	s.Methods("GET").Path("/{name}").HandlerFunc(Name2VolumeHandlerFunc(man.Get))
	s.Methods("DELETE").Path("/{name}").HandlerFunc(NameHandlerFunc(man.Delete))
	s.Methods("POST").Path("/{name}/attach").HandlerFunc(NameHandlerFunc(man.Attach))
	s.Methods("POST").Path("/{name}/detach").HandlerFunc(NameHandlerFunc(man.Detach))

	return r
}

func NameHandlerFunc(f func(name string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		err := f(name)
		switch err {
		case nil:
			r.JSON(w, http.StatusOK, map[string]interface{}{})
		default:
			r.JSON(w, http.StatusBadGateway, err)
		}
	}
}

func Name2VolumeHandlerFunc(f func(name string) (*types.VolumeInfo, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		volume, err := f(name)
		switch err {
		case nil:
			r.JSON(w, http.StatusOK, volume)
		default:
			r.JSON(w, http.StatusBadGateway, err)
		}
	}
}

func Volume2VolumeHandlerFunc(f func(volume *types.VolumeInfo) (*types.VolumeInfo, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		volume0 := new(types.VolumeInfo)
		if err := json.NewDecoder(req.Body).Decode(volume0); err != nil {
			r.JSON(w, http.StatusBadRequest, errors.Wrap(err, "could not parse"))
			return
		}
		volume, err := f(volume0)
		switch err {
		case nil:
			r.JSON(w, http.StatusOK, volume)
		default:
			r.JSON(w, http.StatusBadGateway, err)
		}
	}
}
