package kvstore

import (
    "DemoBlockChain/lib"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	dbm "github.com/tendermint/tm-db"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/kv"
	"github.com/tendermint/tendermint/version"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")

	ProtocolVersion version.Protocol = 0x1
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

//获取kv
func loadState(db dbm.DB) State {
	var state State
	state.db = db
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

//保存kv
func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

//---------------------------------------------------

var _ types.Application = (*Application)(nil)

type Application struct {
	types.BaseApplication

	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication() *Application {
	state := loadState(dbm.NewMemDB())
	return &Application{state: state}
}

func (app *Application) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	return types.ResponseInfo{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCIVersion,
		AppVersion:       ProtocolVersion.Uint64(),
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}
}

// tx is either "key=value" or just arbitrary bytes
func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	var method, key, value []byte
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) == 3 {
		method, key, value = parts[0], parts[1], parts[2]
	} else {
		method, key, value = req.Tx, req.Tx, req.Tx
	}

    lib.Log.Notice(string(method))
	lib.Log.Notice(string(key))
    lib.Log.Notice(string(value))

    switch string(method) {
        case "add":
            // 此处修改 app.state.db.Set(prefixKey(key), value)
            app.state.db.Set(key, value)
            app.state.Size++
        case "modify":
            exist, e := app.state.db.Has(key)
            lib.Log.Notice(exist)
            if e == nil {
                app.state.db.Delete(key)
                app.state.db.Set(key, value)
            }
        case "delete":
            exist, e := app.state.db.Has(key)
            lib.Log.Notice(exist)
            if e == nil {
                app.state.db.Delete(key)
            }
    }

	events := []types.Event{
		{
			Type: "app",
			Attributes: []kv.Pair{
				{Key: []byte("creator"), Value: []byte("Cosmoshi Netowoko")},
				{Key: []byte("key"), Value: key},
			},
		},
	}

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Events: events}
}

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
}

func (app *Application) Commit() types.ResponseCommit {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)

	resp := types.ResponseCommit{Data: appHash}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp
}

// Returns an associated value or nil if missing.
func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
    lib.Log.Notice(reqQuery)
// 	if reqQuery.Prove {
// 	    lib.Log.Notice(string(reqQuery.Data))
	    // 此处修改 value, err := app.state.db.Get(prefixKey(reqQuery.Data))
// 		value, err := app.state.db.Get(reqQuery.Data)
//
// 		if err != nil {
// 			panic(err)
// 		}
// 		if value == nil {
// 			resQuery.Log = "does not exist"
// 		} else {
// 			resQuery.Log = "exists"
// 		}
// 		resQuery.Index = -1 // TODO make Proof return index
// 		resQuery.Key = reqQuery.Data
// 		resQuery.Value = value
// 		resQuery.Height = app.state.Height
//
// 		return resQuery
// 	}
    lib.Log.Notice(string(reqQuery.Path))
    if reqQuery.Path == "" {
        resQuery.Key = reqQuery.Data
        // 此处修改 value, err := app.state.db.Get(prefixKey(reqQuery.Data))
        value, err := app.state.db.Get(reqQuery.Data)

        if err != nil {
            panic(err)
        }
        if value == nil {
            resQuery.Log = "does not exist"
        } else {
            resQuery.Log = "exists"
        }
        resQuery.Value = value
        resQuery.Height = app.state.Height


    }else{
        // 迭代器
        itr, e := app.state.db.Iterator(nil, nil)
        // 查询kv获取对应数据
        var build strings.Builder
        build.WriteString("[")
        for ; itr.Valid(); itr.Next() {
            key := itr.Key()
            value := itr.Value()
            if strings.Index(string(key), reqQuery.Path) != -1 && strings.Index(string(value), string(reqQuery.Data)) != -1 {
                build.WriteString(string(value))
                build.WriteString(",")
            }
        }
        result := build.String()
        result = strings.TrimRight(result, ",")
        result = result + "]"
        lib.Log.Notice(result)
        lib.Log.Notice(e)
        resQuery.Key = reqQuery.Data
        resQuery.Value = []byte(result)
    }

    return resQuery
}
