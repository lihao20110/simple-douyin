package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

//GetCoverImage  生成视频缩略图作为封面并保存
//videoPath:视频的路径
//saveCoverPath:生成的缩略图保存的路径
//frameNum:缩略图所属的帧数,指定要截取视频的哪一帧作为封面,通过调整帧数截取稍微靠后的几秒作为封面
//CoverImageName:缩略图的文件名称
func GetCoverImage(videoPath, saveCoverPath string, frameNum int) (CoverImageName string, err error) {
	buf := bytes.NewBuffer(nil)
	err = ffmpeg.Input(videoPath).Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	img, err := imaging.Decode(buf)
	if err != nil {
		panic(err)
	}
	err = imaging.Save(img, saveCoverPath+".jpeg")
	if err != nil {
		panic(err)
	}
	//成功返回生成的封面图名
	names := strings.Split(saveCoverPath, "/")
	CoverImageName = names[len(names)-1] + ".jpeg"
	return
}
