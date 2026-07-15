// Package mocks 存放由 gomock 生成的 mock 对象。
package mocks

//go:generate mockgen -destination=mock_UserServiceClient.go -package=mocks github.com/Jedeft/demo-micro-base-user/api/protobuf UserServiceClient
