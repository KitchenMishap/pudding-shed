package corereader

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// corereader.CoreReader implements jsonblock.IBlockJsonFetcher
var _ jsonblock.IBlockJsonFetcher = (*CoreReader)(nil) // Check that implements

var latestBlock = int64(0)

type CoreReader struct {
	httpClient http.Client
}

func (cr *CoreReader) getHashHexByHeight(height int64) (string, error) {
	// Talk to core REST
	// You will need the core options -server -rest
	// Also recommended are -txindex -disablewallet
	req := "http://127.0.0.1:8332/rest/blockhashbyheight/"
	req += strconv.Itoa(int(height)) + ".hex"
	resp, err := cr.httpClient.Get(req)

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
	resp, err := cr.httpClient.Get(req)

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
	if height < latestBlock {
		print("Height ", height, " is less than latest block ", latestBlock)
	}

	hashHexString, err := cr.getHashHexByHeight(height)
	if err != nil {
		return nil, err
	}

	req := "http://127.0.0.1:8332/rest/block/"
	req += hashHexString
	req += ".json"

	resp, err := cr.httpClient.Get(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	latestBlock = height

	return jsonBytes, nil
}
