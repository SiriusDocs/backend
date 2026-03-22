package services

import (
	"context"
	"io"

	temp "github.com/SiriusDocs/protos/gen/go/templates"
)

type TempService interface {
    UploadFile(ctx context.Context, filename string, fileReader io.Reader) (string, error)
    CheckStatus(ctx context.Context, req *temp.StatusRequest)(*temp.StatusResponse, error)
}

type tempService struct {
    tempClient temp.TempClient
}

func NewTempService(client temp.TempClient) TempService {
    return &tempService{
        tempClient: client,
    }
}

func (t *tempService) UploadFile(ctx context.Context, filename string, fileReader io.Reader) (string, error) {
   stream, err := t.tempClient.UploadAndParse(ctx)
    if err != nil {
        return "", err
    }

    // отправляем метаданные (первое сообщение)
    err = stream.Send(&temp.UploadRequest{
        Data: &temp.UploadRequest_Info{
            Info: &temp.FileInfo{Filename: filename},
        },
    })
    if err != nil {
        return "", err
    }

    // читаем файл кусками и отправляем в стрим
    buffer := make([]byte, 32*1024) // 32KB
    for {
        n, err := fileReader.Read(buffer)
        if err == io.EOF {
            break // конец файла
        }
        if err != nil {
            return "", err
        }

        err = stream.Send(&temp.UploadRequest{
            Data: &temp.UploadRequest_ChunkData{
                ChunkData: buffer[:n],
            },
        })
        if err != nil {
            return "", err
        }
    }

    // закрываем стрим и ждем ответ от сервера
    res, err := stream.CloseAndRecv()
    if err != nil {
        return "", err
    }

    return res.TaskId, nil
}

func (t *tempService) CheckStatus(ctx context.Context, req *temp.StatusRequest)(*temp.StatusResponse, error) {
    return t.tempClient.CheckStatus(ctx, req)
}