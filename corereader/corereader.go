package corereader

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// corereader.CoreReader implements jsonblock.IBlockJsonFetcher
var _ jsonblock.IBlockJsonFetcher = (*CoreReader)(nil) // Check that implements

var theTransport1 = http.Transport{ResponseHeaderTimeout: time.Hour, MaxConnsPerHost: 15, DisableKeepAlives: true}
var theClient1 = http.Client{Transport: &theTransport1}
var theTransport2 = http.Transport{ResponseHeaderTimeout: time.Hour, MaxConnsPerHost: 15, DisableKeepAlives: true}
var theClient2 = http.Client{Transport: &theTransport1}

type CoreReader struct {
	latestBlock int64
	client      *http.Client
}

func (cr *CoreReader) getHashHexByHeight(height int64) (string, error) {
	// Talk to core REST
	// You will need the core options -server -rest
	// Also recommended are -txindex -disablewallet -rpcthreads=15
	req := "http://127.0.0.1:8332/rest/blockhashbyheight/"
	req += strconv.Itoa(int(height)) + ".hex"

	var resp *http.Response
	var err error
	resp, err = cr.client.Get(req)

	if err != nil {
		println(err.Error())
		println("getHashHexByHeight(): Could not GET from local bitcoin REST server")
		println("Are you sure Bitcoin Core is running, with the correct parameters?")
		println("Recommend: bitcoin-qt.exe -txindex -disablewallet -server -rest")
		return "", err
	}
	defer resp.Body.Close()
	bodyout, _ := io.ReadAll(resp.Body)
	hex := strings.TrimSpace(string(bodyout)) // Strip trailing newline
	return hex, nil
}

// functions to implement jsonblock.IBlockJsonFetcher

func (cr *CoreReader) CountBlocks() (int64, error) {
	req := "http://127.0.0.1:8332/rest/chaininfo.json"
	var resp *http.Response
	var err error
	resp, err = cr.client.Get(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	jsonBytes, _ := io.ReadAll(resp.Body)

	var info jsonChainInfoForBlocksCount
	err = json.Unmarshal(jsonBytes, &info)
	if err != nil {
		return -1, err
	}

	return info.Blocks, nil
}

func (cr *CoreReader) FetchBlockJsonBytes(height int64) ([]byte, error) {
	if height < cr.latestBlock {
		print("Height ", height, " is less than latest block ", cr.latestBlock)
	}

	hashHexString, err := cr.getHashHexByHeight(height)
	if err != nil {
		return nil, err
	}

	req := "http://127.0.0.1:8332/rest/block/"
	req += hashHexString
	req += ".json"

	var resp *http.Response
	resp, err = cr.client.Get(req)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	cr.latestBlock = height

	return jsonBytes, nil
}
