package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/spf13/cobra"
	"github.com/tonycai653/iqshell/qshell"
	"io/ioutil"
	"os"
)

var qUploadCmd = &cobra.Command{
	Use:   "qupload <quploadConfigFile>",
	Short: "Batch upload files to the qiniu bucket",
	Args:  cobra.ExactArgs(1),
	Run:   QiniuUpload,
}

var (
	successFname   string
	failureFname   string
	overwriteFname string
	upthreadCount  int64
	uploadConfig   qshell.UploadConfig
)

func init() {
	qUploadCmd.Flags().StringVarP(&successFname, "success-list", "s", "", "upload success (all) file list")
	qUploadCmd.Flags().StringVarP(&failureFname, "failure-list", "f", "", "upload failure file list")
	qUploadCmd.Flags().StringVarP(&overwriteFname, "overwrite-list", "w", "", "upload success (overwrite) file list")
	qUploadCmd.Flags().Int64VarP(&upthreadCount, "worker", "c", 1, "worker count")
	RootCmd.AddCommand(qUploadCmd)
}

func parseUploadConfigFile(uploadConfigFile string, uploadConfig *qshell.UploadConfig) (err error) {
	//read upload config
	if uploadConfigFile == "" {
		err = fmt.Errorf("config filename is empty")
		return
	}
	fp, oErr := os.Open(uploadConfigFile)
	if oErr != nil {
		err = fmt.Errorf("Open upload config file ``%s`: %v\n", uploadConfigFile, oErr)
		return
	}
	defer fp.Close()

	configData, rErr := ioutil.ReadAll(fp)
	if rErr != nil {
		err = fmt.Errorf("Read upload config file `%s`: %v\n", uploadConfigFile, rErr)
		return
	}
	uErr := json.Unmarshal(configData, uploadConfig)
	if uErr != nil {
		err = fmt.Errorf("Parse upload config file `%s`: %v\n", uploadConfigFile, uErr)
		return
	}
	return
}

func QiniuUpload(cmd *cobra.Command, params []string) {

	configFile := params[0]

	pErr := parseUploadConfigFile(configFile, &uploadConfig)
	if pErr != nil {
		logs.Error(fmt.Sprintf("parse config file: %s: %v\n", configFile, pErr))
		os.Exit(qshell.STATUS_HALT)
	}

	if uploadConfig.FileType != 1 && uploadConfig.FileType != 0 {
		logs.Error("Wrong Filetype, It should be 0 or 1 ")
		os.Exit(qshell.STATUS_HALT)
	}

	srcFileInfo, err := os.Stat(uploadConfig.SrcDir)
	if err != nil {
		logs.Error("Upload config error for parameter `SrcDir`,", err)
		os.Exit(qshell.STATUS_HALT)
	}

	if !srcFileInfo.IsDir() {
		logs.Error("Upload src dir should be a directory")
		os.Exit(qshell.STATUS_HALT)
	}

	//upload
	if upthreadCount < qshell.MIN_UPLOAD_THREAD_COUNT || upthreadCount > qshell.MAX_UPLOAD_THREAD_COUNT {
		logs.Info("Tip: you can set <ThreadCount> value between %d and %d to improve speed\n",
			qshell.MIN_UPLOAD_THREAD_COUNT, qshell.MAX_UPLOAD_THREAD_COUNT)

		if upthreadCount < qshell.MIN_UPLOAD_THREAD_COUNT {
			upthreadCount = qshell.MIN_UPLOAD_THREAD_COUNT
		} else if upthreadCount > qshell.MAX_UPLOAD_THREAD_COUNT {
			upthreadCount = qshell.MAX_UPLOAD_THREAD_COUNT
		}
	}

	fileExporter, fErr := qshell.NewFileExporter(successFname, failureFname, overwriteFname)
	if fErr != nil {
		logs.Error("initialize fileExporter: ", fErr)
		os.Exit(qshell.STATUS_HALT)
	}
	qshell.QiniuUpload(int(upthreadCount), &uploadConfig, fileExporter)
}