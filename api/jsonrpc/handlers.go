//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package jsonrpc

import (
	"errors"

	"github.com/eclipse/che-machine-exec/api/events"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/cfg"
	"github.com/sirupsen/logrus"

	"strconv"

	jsonrpc "github.com/eclipse/che-go-jsonrpc"
	"github.com/eclipse/che-machine-exec/exec"
)

const (
	// BearerTokenAttr attribute name.
	BearerTokenAttr = "bearerToken"
)

type IdParam struct {
	Id int `json:"id"`
}

type OperationResult struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
}

type ResizeParam struct {
	Id   int  `json:"id"`
	Cols uint `json:"cols"`
	Rows uint `json:"rows"`
}

type EmptyParam struct {
}

var (
	execManager = exec.GetExecManager()
)

func jsonRpcCreateExec(tunnel *jsonrpc.Tunnel, params interface{}, t jsonrpc.RespTransmitter) {
	machineExec := params.(*model.MachineExec)
	if cfg.UseBearerToken {
		if token, ok := tunnel.Attributes[BearerTokenAttr]; ok && len(token) > 0 {
			machineExec.BearerToken = token
		} else {
			err := errors.New("Bearer token should not be an empty")
			logrus.Errorf(err.Error())
			t.SendError(jsonrpc.NewArgsError(err))
			return
		}
	}

	id, err := execManager.Create(machineExec)

	healthWatcher := exec.NewHealthWatcher(machineExec, events.EventBus, execManager)
	healthWatcher.CleanUpOnExitOrError()

	if err != nil {
		logrus.Errorf("Unable to initialize terminal. Cause: %s", err.Error())
		t.SendError(jsonrpc.NewArgsError(err))
		return
	}

	if id == -1 {
		logrus.Errorln("A container where it's possible to initialize terminal is not found")
		t.SendError(jsonrpc.NewArgsError(errors.New("A container where it's possible to initialize terminal is not found")))
		return
	}

	t.Send(id)
}

func jsonRpcCheckExec(_ *jsonrpc.Tunnel, params interface{}, t jsonrpc.RespTransmitter) {
	idParam := params.(*IdParam)

	id, err := execManager.Check(idParam.Id)
	if err != nil {
		t.SendError(jsonrpc.NewArgsError(err))
	}

	t.Send(id)
}

func jsonRpcResizeExec(_ *jsonrpc.Tunnel, params interface{}) (interface{}, error) {
	resizeParam := params.(*ResizeParam)

	if err := execManager.Resize(resizeParam.Id, resizeParam.Cols, resizeParam.Rows); err != nil {
		return nil, jsonrpc.NewArgsError(err)
	}

	return &OperationResult{
		Id: resizeParam.Id, Text: "Exec with id " + strconv.Itoa(resizeParam.Id) + "  was successfully resized",
	}, nil
}

// func jsonRpcCreateKubeConfigExec(tunnel *jsonrpc.Tunnel, _ interface{}, t jsonrpc.RespTransmitter) {
// 	config := ""

// 	infoExecCreator := exec_info.NewKubernetesInfoExecCreator(exec.GetNamespace(), k8sAPI.GetClient().Core(), k8sAPI.GetConfig())

// 	// Find the correct container or whatever then make another exec info request to find where to store it
// 	// then store it basically like what serheii had

// 	if cfg.UseBearerToken {
// 		if token, ok := tunnel.Attributes[BearerTokenAttr]; ok && len(token) > 0 {
// 			config = kubeconfig.CreateKubeConfig(token)
// 			infoExec = infoExecCreator.CreateInfoExec([]string{"sh", "-c", "echo $KUBECONFIG"}, containerInfo)
// 			if err := infoExec.Start(); err != nil {
// 				logrus.Debugf("Error is not available in %s/%s. Error: %s", containerInfo.PodName, containerInfo.ContainerName, err.Error())
// 				return
// 			}
// 			kubeconfigLocation := infoExec.GetOutput()

// 			logrus.Debugf("Creating /tmp/.kube in %s/%s", containerInfo.PodName, containerInfo.ContainerName)
// 			infoExec := infoExecCreator.CreateInfoExec([]string{"sh", "-c", "mkdir -p " + kubeconfigLocation}, containerInfo)
// 			if err := infoExec.Start(); err != nil {
// 				logrus.Debugf("Error is not available in %s/%s. Error: %s", containerInfo.PodName, containerInfo.ContainerName, err.Error())
// 				return
// 			}

// 			logrus.Debugf("Writing token in /tmp/.kube/token in %s/%s", containerInfo.PodName, containerInfo.ContainerName)
// 			infoExec = infoExecCreator.CreateInfoExec([]string{"sh", "-c", "echo " + config + " > " + kubeconfigLocation}, containerInfo)
// 			if err := infoExec.Start(); err != nil {
// 				logrus.Debugf("Error is not available in %s/%s. Error: %s", containerInfo.PodName, containerInfo.ContainerName, err.Error())
// 				return
// 			}

// 			t.Send(config)
// 		} else {
// 			createTunneledError(errors.New("Bearer token should not be an empty"), t)
// 		}
// 	} else {
// 		createTunneledError(errors.New("Kubeconfig environment variable was set but bearer token not used"), t)
// 	}
// }

// func createTunneledError(err error, t jsonrpc.RespTransmitter) {
// 	logrus.Errorf(err.Error())
// 	t.SendError(jsonrpc.NewArgsError(err))
// }
