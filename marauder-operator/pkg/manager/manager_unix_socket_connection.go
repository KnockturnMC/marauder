package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"path"
	"strings"
	"time"

	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ErrServerWithoutManagementSocket is returned if a server does not have a management socket assigned in DB.
var ErrServerWithoutManagementSocket = errors.New("server does not have management socket")

// ErrMessageTooLarge the message to send is too large.
var ErrMessageTooLarge = errors.New("failed to send message as it is too large")

// TypeURLPrefix prefixed for any message.
const TypeURLPrefix = "knockturnmc.com/"

// ExchangeManagementMessage performs a synchronous request-response exchange over a Unix domain socket.
// It opens a connection to the server's management socket, sends the outgoing Protobuf message,
// and blocks until the incoming response is received and unmarshaled.
//
// The communication follows a length-prefixed protocol:
// 1. A 4-byte big-endian integer indicating the message length.
// 2. The marshaled Protobuf bytes.
func (d DockerBasedManager) ExchangeManagementMessage(
	ctx context.Context,
	model networkmodel.ServerModel,
	outgoing proto.Message,
	incoming proto.Message,
) error {
	if model.ManagementSocketPath == "" {
		return ErrServerWithoutManagementSocket
	}

	conn, err := openUnixSocketConn(d, model)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	deadline, ok := ctx.Deadline()
	if ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set read/write timeout for connection: %w", err)
		}
	}

	defer func() { utils.Swallow(conn.Close()) }()

	if err := writeProtoMessage(conn, outgoing); err != nil {
		return fmt.Errorf("failed to write proto message: %w", err)
	}

	if err := readProtoMessage(conn, incoming); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return nil
}

func openUnixSocketConn(d DockerBasedManager, model networkmodel.ServerModel) (net.Conn, error) {
	location, err := d.computeServerFolderLocation(model)
	if err != nil {
		return nil, fmt.Errorf("failed to compute socket location: %w", err)
	}

	socketPathWithoutMount, _ := strings.CutPrefix(model.ManagementSocketPath, ServerMountTarget)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	dial, err := dialer.Dial("unix", path.Join(location, socketPathWithoutMount))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to management socket: %w", err)
	}

	return dial, nil
}

// writeProtoMessage marshals a Protobuf message and writes it to the provided writer.
func writeProtoMessage[T proto.Message](writer io.Writer, outgoing T) error {
	payload, err := anypb.New(outgoing)
	if err != nil {
		return fmt.Errorf("failed to marshal outgoing into any: %w", err)
	}

	payload.TypeUrl = TypeURLPrefix + string(outgoing.ProtoReflect().Descriptor().FullName())

	marshal, err := proto.Marshal(marauderpb.Message_builder{Payload: payload}.Build())
	if err != nil {
		return fmt.Errorf("failed to marshal outgoing message: %w", err)
	}

	messageLen := len(marshal)
	if messageLen > math.MaxInt32 {
		return fmt.Errorf("%d > %d: %w", messageLen, math.MaxInt32, ErrMessageTooLarge)
	}

	if _, err := writer.Write(utils.IntToByteSlice(int32(messageLen))); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}

	if _, err := writer.Write(marshal); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

// readProtoMessage reads a length-prefixed Protobuf message from the provided reader.
func readProtoMessage[T proto.Message](reader io.Reader, incoming T) error {
	size, err := utils.ByteSliceToInt(reader)
	if err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	messageBytes := make([]byte, size)
	if _, err := io.ReadFull(reader, messageBytes); err != nil {
		return fmt.Errorf("failed to read message from stream: %w", err)
	}

	var message marauderpb.Message
	if err := proto.Unmarshal(messageBytes, &message); err != nil {
		return fmt.Errorf("failed to unmarshal incoming message: %w", err)
	}

	if err := message.GetPayload().UnmarshalTo(incoming); err != nil {
		return fmt.Errorf("failed to unmarshal expected payload: %w", err)
	}

	return nil
}
