package cmds

import (
	"bufio"
	"encoding/json"
	"log"

	"github.com/mdlayher/vsock"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/constants"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/messages"
	"github.com/zxsdotch/aws-nitro-enclave-experiments/foobar-shared/utils"
)

func sendRequest(req messages.FoobarRequest) (messages.FoobarResponse, []byte) {
	log.Printf("Connecting to vsock (cid=%d, port=%d)\n", constants.ENCLAVE_CID, constants.ENCLAVE_LISTENING_PORT)
	conn, err := vsock.Dial(constants.ENCLAVE_CID, constants.ENCLAVE_LISTENING_PORT, nil)
	utils.PanicOnErr(err)
	defer conn.Close()

	log.Printf("Send: %v", req)
	msgBytes, err := json.Marshal(req)
	utils.PanicOnErr(err)

	conn.Write(msgBytes)
	conn.Write([]byte{'\n'})
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	var resp messages.FoobarResponse
	json.Unmarshal(scanner.Bytes(), &resp)
	log.Printf("Recv: %v", resp)

	if resp.Error != nil {
		log.Panicf("enclave error: %s\n", *resp.Error)
	}

	return resp, msgBytes
}
