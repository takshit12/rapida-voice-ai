package internal_grpc

import (
	"context"

	internal_audio "github.com/rapidaai/api/assistant-api/internal/audio"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_text "github.com/rapidaai/api/assistant-api/internal/text"
	"github.com/rapidaai/protos"
	"google.golang.org/grpc"
)

type unidirectionalStreamer struct {
	server grpc.BidiStreamingServer[protos.AssistantMessagingRequest, protos.AssistantMessagingResponse]
}

func NewGrpcUnidirectionalStreamer(
	server protos.TalkService_AssistantTalkServer) internal_streamers.Streamer {
	return &unidirectionalStreamer{
		server: server,
	}
}

func (uds *unidirectionalStreamer) Context() context.Context {
	return uds.server.Context()
}

func (uds *unidirectionalStreamer) Recv() (*protos.AssistantMessagingRequest, error) {
	return uds.server.Recv()
}

// Send sends an output value to the stream.
// It returns an error if the send operation fails.
func (uds *unidirectionalStreamer) Send(out *protos.AssistantMessagingResponse) error {
	return uds.server.Send(out)
}

func (extl *unidirectionalStreamer) Config() *internal_streamers.StreamAttribute {
	return internal_streamers.NewStreamAttribute(
		internal_streamers.NewStreamConfig(internal_audio.NewLinear16khzMonoAudioConfig(),
			&internal_text.TextConfig{
				Charset: "UTF-8",
			},
		), internal_streamers.NewStreamConfig(internal_audio.NewLinear16khzMonoAudioConfig(),
			&internal_text.TextConfig{
				Charset: "UTF-8",
			},
		))
}
