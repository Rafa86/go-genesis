// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type tableInfo struct {
	Name  string `json:"name"`
	Count string `json:"count"`
}

type tablesResult struct {
	Count int64       `json:"count"`
	List  []tableInfo `json:"list"`
}

func tablesHandler(w http.ResponseWriter, r *http.Request) {
	form := &paginatorForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	table := client.Prefix() + `_tables`
	count, err := model.GetRecordsCountTx(nil, table)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting records count from tables")
		errorResponse(w, err)
		return
	}

	// TODO: перенести в модели
	list, err := model.GetAll(`select name from "`+table+`" order by name`+
		fmt.Sprintf(` offset %d `, form.Offset), form.Limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting names from tables")
		errorResponse(w, err)
		return
	}

	result := &tablesResult{
		Count: count,
		List:  make([]tableInfo, len(list)),
	}

	for i, item := range list {
		var maxid int64
		result.List[i].Name = item[`name`]
		fullname := client.Prefix() + `_` + item[`name`]
		if item[`name`] == `keys` || item[`name`] == `members` {
			err = model.DBConn.Table(fullname).Count(&maxid).Error
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting count from table")
			}
		} else {
			maxid, err = model.GetNextID(nil, fullname)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id from table")
			}
			maxid--
		}
		if err != nil {
			errorResponse(w, err)
			return
		}
		result.List[i].Count = converter.Int64ToStr(maxid)
	}

	jsonResponse(w, result)
}
