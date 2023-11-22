package fungo

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/paololu/funplugin/fungo/protoGen"
	jsoniter "github.com/json-iterator/go"
)

// replace with third-party json library to improve performance
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// functionGRPCClient runs on the host side, it implements FuncCaller interface
type functionGRPCClient struct {
	client protoGen.DebugTalkClient
}

func (m *functionGRPCClient) GetNames() ([]string, error) {
	logger.Debug("gRPC_client GetNames() start")
	resp, err := m.client.GetNames(context.Background(), &protoGen.Empty{})
	if err != nil {
		logger.Error("gRPC_client GetNames() failed", "error", err)
		return nil, err
	}
	logger.Debug("gRPC_client GetNames() success")
	return resp.Names, nil
}

func (m *functionGRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	logger.Info("gRPC_client Call() start", "funcName", funcName, "funcArgs", funcArgs)

	funcArgBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() funcArgs")
	}
	req := &protoGen.CallRequest{
		Name: funcName,
		Args: funcArgBytes,
	}

	response, err := m.client.Call(context.Background(), req)
	if err != nil {
		logger.Error("gRPC_client Call() failed",
			"funcName", funcName,
			"funcArgs", funcArgs,
			"error", err,
		)
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(response.Value, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() response")
	}
	logger.Info("gRPC_client Call() success", "result", resp)
	return resp, nil
}

// Here is the gRPC server that functionGRPCClient talks to.
type functionGRPCServer struct {
	protoGen.UnimplementedDebugTalkServer
	Impl IFuncCaller
}

func (m *functionGRPCServer) GetNames(ctx context.Context, req *protoGen.Empty) (*protoGen.GetNamesResponse, error) {
	logger.Debug("gRPC_server GetNames() start")
	v, err := m.Impl.GetNames()
	if err != nil {
		logger.Error("gRPC_server GetNames() failed", "error", err)
		return nil, err
	}
	logger.Debug("gRPC_server GetNames() success")
	return &protoGen.GetNamesResponse{Names: v}, nil
}

func (m *functionGRPCServer) Call(ctx context.Context, req *protoGen.CallRequest) (*protoGen.CallResponse, error) {
	logger.Debug("gRPC_server Call() start")

	var funcArgs []interface{}
	if err := json.Unmarshal(req.Args, &funcArgs); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() funcArgs")
	}

	v, err := m.Impl.Call(req.Name, funcArgs...)
	if err != nil {
		logger.Error("gRPC_server Call() failed", "req", req, "error", err)
		return nil, err
	}

	value, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() response")
	}
	logger.Debug("gRPC_server Call() success")
	return &protoGen.CallResponse{Value: value}, nil
}

// GRPCPlugin implements hashicorp's plugin.GRPCPlugin.
type GRPCPlugin struct {
	plugin.Plugin
	Impl IFuncCaller
}

func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	protoGen.RegisterDebugTalkServer(s, &functionGRPCServer{Impl: p.Impl})
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &functionGRPCClient{client: protoGen.NewDebugTalkClient(c)}, nil
}
