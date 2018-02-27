package modes

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/GenesisKernel/go-genesis/packages/api"
	"github.com/julienschmidt/httprouter"
	"github.com/rpoletaev/supervisord/process"
	log "github.com/sirupsen/logrus"
)

// AddVDEMasterHandlers add specific handlers to router
func (mode *VDEMaster) registerHandlers(router *httprouter.Router) {
	api.MethodRoute(router, http.MethodPost, "/vde/create", "", true, mode.createVDEHandler)
	api.MethodRoute(router, http.MethodPost, "/vde/start", "", true, mode.startVDEHandler)
	api.MethodRoute(router, http.MethodPost, "/vde/stop", "", true, mode.stopVDEHandler)
	api.MethodRoute(router, http.MethodPost, "/vde/delete", "", true, mode.deleteVDEHandler)
	api.MethodRoute(router, http.MethodGet, "/vde", "", true, mode.listVDEHandler)
}

func (mode *VDEMaster) createVDEHandler(w http.ResponseWriter, r *http.Request, data *api.ApiData, logger *log.Entry) error {
	name := r.FormValue("name")
	if len(name) == 0 {
		return api.ErrorAPI(w, fmt.Errorf("name is empty"), http.StatusBadRequest)
	}

	user := r.FormValue("dbUser")
	password := r.FormValue("dbPassword")

	if err := mode.CreateVDE(name, user, password); err != nil {
		return api.ErrorAPI(w, err, http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "VDE '%s' created", name)
	return nil
}

func (mode *VDEMaster) startVDEHandler(w http.ResponseWriter, r *http.Request, data *api.ApiData, logger *log.Entry) error {
	name := r.FormValue("name")
	if len(name) == 0 {
		return api.ErrorAPI(w, fmt.Errorf("name is empty"), http.StatusBadRequest)
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		return api.ErrorAPI(w, fmt.Sprintf("process '%s' not found", name), http.StatusNotFound)
	}

	state := proc.GetState()
	if state == process.STOPPED ||
		state == process.EXITED ||
		state == process.FATAL {
		proc.Start(true)
		fmt.Fprintf(w, "VDE '%s' is started", name)
		return nil
	}

	return api.ErrorAPI(w, fmt.Errorf("VDE '%s' is %s", name, state), http.StatusBadRequest)
}

func (mode *VDEMaster) stopVDEHandler(w http.ResponseWriter, r *http.Request, data *api.ApiData, logger *log.Entry) error {
	name := r.FormValue("name")
	if len(name) == 0 {
		return api.ErrorAPI(w, fmt.Errorf("name is empty"), http.StatusBadRequest)
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		return api.ErrorAPI(w, fmt.Errorf("process '%s' not found", name), http.StatusNotFound)
	}

	state := proc.GetState()
	if state == process.RUNNING ||
		state == process.STARTING {
		proc.Stop(true)
		fmt.Fprintf(w, "VDE '%s' is stoped", name)
		return nil
	}

	return api.ErrorAPI(w, fmt.Errorf("VDE '%s' is %s", name, state), http.StatusBadRequest)
}

func (mode *VDEMaster) deleteVDEHandler(w http.ResponseWriter, r *http.Request, data *api.ApiData, logger *log.Entry) error {
	name := r.FormValue("name")
	if len(name) == 0 {
		return api.ErrorAPI(w, fmt.Errorf("name is empty"), http.StatusBadRequest)
	}

	proc := mode.processes.Find(name)
	if proc == nil {
		return api.ErrorAPI(w, fmt.Errorf("process '%s' not found", name), http.StatusNotFound)
	}

	proc.Stop(true)
	return nil
}

func (mode *VDEMaster) listVDEHandler(w http.ResponseWriter, r *http.Request, data *api.ApiData, logger *log.Entry) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(mode.ListProcess()); err != nil {
		return api.ErrorAPI(w, err, http.StatusInternalServerError)
	}

	return nil
}