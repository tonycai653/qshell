package cmd

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/spf13/cobra"
	"github.com/tonycai653/iqshell/qshell"
	"os"
)

var qUpload2Cmd = &cobra.Command{
	Use:   "qupload2",
	Short: "Batch upload files to the qiniu bucket",
	Run:   QiniuUpload2,
}

var (
	// defined in qupload.go
	// uploadConfig   qshell.UploadConfig
	up2threadCount int64
)

func init() {
	qUpload2Cmd.Flags().Int64Var(&up2threadCount, "thread-count", 0, "multiple thread count")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.SrcDir, "src-dir", "", "src dir to upload")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.FileList, "file-list", "", "file list to upload")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.Bucket, "bucket", "", "bucket")
	qUpload2Cmd.Flags().Int64Var(&uploadConfig.PutThreshold, "put-threshold", 0, "chunk upload threshold")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.KeyPrefix, "key-prefix", "", "key prefix prepended to dest file key")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.IgnoreDir, "ignore-dir", false, "ignore the dir in the dest file key")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.Overwrite, "overwrite", false, "overwrite the file of same key in bucket")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.CheckExists, "check-exists", false, "check file key whether in bucket before upload")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.CheckHash, "check-hash", false, "check hash")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.CheckSize, "check-size", false, "check file size")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.SkipFilePrefixes, "skip-file-prefixes", "", "skip files with these file prefixes")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.SkipPathPrefixes, "skip-path-prefixes", "", "skip files with these relative path prefixes")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.SkipFixedStrings, "skip-fixed-strings", "", "skip files with the fixed string in the name")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.SkipSuffixes, "skip-suffixes", "", "skip files with these suffixes")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.UpHost, "up-host", "", "upload host")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.BindUpIp, "bind-up-ip", "", "upload host ip to bind")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.BindRsIp, "bind-rs-ip", "", "rs host ip to bind")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.BindNicIp, "bind-nic-ip", "", "local network interface card to bind")
	qUpload2Cmd.Flags().BoolVar(&uploadConfig.RescanLocal, "rescan-local", false, "rescan local dir to upload newly add files")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.LogFile, "log-file", "", "log file")
	qUpload2Cmd.Flags().StringVar(&uploadConfig.LogLevel, "log-level", "info", "log level")
	qUpload2Cmd.Flags().IntVar(&uploadConfig.LogRotate, "log-rotate", 1, "log rotate days")
	qUpload2Cmd.Flags().IntVar(&uploadConfig.FileType, "file-type", 0, "set storage file type")
	qUpload2Cmd.Flags().StringVar(&successFname, "success-list", "", "upload success file list")
	qUpload2Cmd.Flags().StringVar(&failureFname, "failure-list", "", "upload failure file list")
	qUpload2Cmd.Flags().StringVar(&overwriteFname, "overwrite-list", "", "upload success (overwrite) file list")

	RootCmd.AddCommand(qUpload2Cmd)
}

func QiniuUpload2(cmd *cobra.Command, params []string) {

	//check params
	if uploadConfig.SrcDir == "" {
		fmt.Println("Upload config no `--src-dir` specified")
		os.Exit(qshell.STATUS_HALT)
	}

	if uploadConfig.Bucket == "" {
		fmt.Println("Upload config no `--bucket` specified")
		os.Exit(qshell.STATUS_HALT)
	}

	if _, err := os.Stat(uploadConfig.SrcDir); err != nil {
		logs.Error("Upload config `SrcDir` not exist error,", err)
		os.Exit(qshell.STATUS_HALT)
	}

	if up2threadCount < qshell.MIN_UPLOAD_THREAD_COUNT || up2threadCount > qshell.MAX_UPLOAD_THREAD_COUNT {
		fmt.Printf("Tip: you can set <ThreadCount> value between %d and %d to improve speed\n",
			qshell.MIN_UPLOAD_THREAD_COUNT, qshell.MAX_UPLOAD_THREAD_COUNT)

		if up2threadCount < qshell.MIN_UPLOAD_THREAD_COUNT {
			up2threadCount = qshell.MIN_UPLOAD_THREAD_COUNT
		} else if up2threadCount > qshell.MAX_UPLOAD_THREAD_COUNT {
			up2threadCount = qshell.MAX_UPLOAD_THREAD_COUNT
		}
	}
	if uploadConfig.FileType != 1 && uploadConfig.FileType != 0 {
		logs.Error("Wrong Filetype, It should be 0 or 1 ")
		os.Exit(qshell.STATUS_HALT)
	}

	fileExporter, fErr := qshell.NewFileExporter(successFname, failureFname, overwriteFname)
	if fErr != nil {
		logs.Error("initialize fileExporter: %v\n", fErr)
		os.Exit(qshell.STATUS_HALT)
	}
	qshell.QiniuUpload(int(up2threadCount), &uploadConfig, fileExporter)
}