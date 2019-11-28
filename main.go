package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
)

type arg struct {
	// 処理対象のファイル
	targetFile string

	// 分割カウントする際の分割数
	splitNum int

	// 同時実行するスレッド(の最大)数
	maxThreads int

	// ファイル読み込み用Bufferのサイズ
	buffersize int
}

var (
	args arg
)

func init() {
	// setup flag help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf("%s -f TARGETFILE [options] [glog options]", os.Args[0]))
		flag.PrintDefaults()
	}

	// glog(の規定値)を設定
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "0")

	// get option(Flag)
	flag.StringVar(&args.targetFile, "f", "", "(Main) Target File.")
	flag.IntVar(&args.splitNum, "s", 2, "Num of File split.")
	flag.IntVar(&args.maxThreads, "t", 2, "Max Num of Threads.")
	flag.IntVar(&args.buffersize, "b", 1024*1024, "Size of ReadBuffer(default=1024*1024).")
}

func countMultiThread(filename string, splitNum int, maxThreads int, buffersize int) (int, error) {
	fp, err := os.OpenFile(filename, 0, 0)
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	// ファイルのバイト数(を取得するためのオブジェクト)を取得
	fileinfo, err := fp.Stat()
	if err != nil {
		return 0, err
	}

	// ファイルのバイト数を取得
	fsize := int(fileinfo.Size())

	glog.V(1).Infof("FileSize   : %10d byte", fsize)
	glog.V(1).Infof("Read buffer: %10d byte", buffersize)
	glog.V(1).Infof("Max Threads: %d", maxThreads)
	glog.V(1).Infof("Split Num  : %d", splitNum)

	// buffersizeの単位で何回読み込みができるかを算出。
	var readCountTotal int = int(math.Trunc(float64(fsize) / float64(buffersize)))

	// あまりがあった場合、読み込み回数に1を加算
	if fsize-(readCountTotal*buffersize) > 0 {
		readCountTotal++
	}

	// 各スレッド(goroutine)に渡す読み込み開始位置
	var byteOffset int64 = 0

	// 終了待機用グループを初期化
	wg := &sync.WaitGroup{}

	// goroutineの同時実行数を制限するためのチャンネル
	jc := make(chan interface{}, maxThreads)
	defer close(jc)

	// 各goroutineの行数カウント結果を受け渡すチャンネル
	countaCh := make(chan int, maxThreads)

	// 各goroutineの終了待ち受けgoroutineから、
	// main処理に集計結果を返すためのチャンネル
	result := make(chan int)
	defer close(result)

	// 結果受信用goroutineを起動
	// 終了条件はclose(countaCh)
	go func(countaCh <-chan int) {
		cAll := 0
		for c := range countaCh {
			cAll += c

			glog.V(2).Infof("[receiver] receive: %d\n", c)
		}

		result <- cAll
	}(countaCh)

	// 個別の行数カウントgoroutineを起動(するためのループ)
	for i := 0; i < splitNum; i++ {
		// countLinesInThread内で、何回buffer読み出しを行うか
		eachReadCount := int(math.Trunc(float64(readCountTotal+i) / float64(splitNum)))

		jc <- true

		wg.Add(1)

		// 個別の行数カウントgoroutineを起動
		go countWorker(filename, eachReadCount, byteOffset, buffersize, wg, jc, countaCh)

		// 読み込みオフセットを進める
		byteOffset += int64(eachReadCount * buffersize)
	}

	wg.Wait()
	close(countaCh)

	return <-result, nil
}

func countWorker(filename string, eachReadCount int, byteOffset int64, buffersize int,
	wg *sync.WaitGroup, jc <-chan interface{}, counta chan<- int) {
	var c int = 0

	defer func() {
		// 無名関数は中から定義元スコープの変数にアクセスできるため。
		counta <- c
		wg.Done()
		<-jc
	}()

	glog.V(2).Infof("[countWorker] start (offset: %d, read size: %d)\n", byteOffset, eachReadCount*buffersize)

	// 対象ファイルを再度開く
	// 外側のファイルハンドラを使用すると、seekと順次読み出しのカーソルがおかしくなる
	fp, err := os.OpenFile(filename, 0, 0)
	if err != nil {
		return
	}
	defer fp.Close()

	// 指定された読み込み開始位置まで移動
	_, err = fp.Seek(byteOffset, 0)
	if err != nil {
		return
	}

	buf := make([]byte, buffersize)

	// 開始位置から、buffersizeづつバイト列を読み込んでbufに代入
	for j := 0; j < eachReadCount; j++ {
		n, err := fp.Read(buf)
		if n == 0 {
			return
		}

		if err != nil {
			// Readエラー処理
			return
		}

		// Bufferの中身を走査するためのオフセット
		of := 0

		// Buffer内の改行を数える
		for {
			// nでサイズを指定しているのは、bufを使いまわしているから
			// (n < len(buf)の場合、前回の読み込み結果が残っている)
			index := bytes.IndexAny(buf[of:n], "\n")
			if index == -1 {
				break
			}

			// (改行の)カウンタをインクリメント
			c++

			// 発見位置+1までオフセットを進める
			of += index + 1
		}
	}
}

func main() {
	flag.Parse()

	glog.V(1).Infof("Start")

	startTime := time.Now()

	linenum, _ := countMultiThread(args.targetFile, args.splitNum, args.maxThreads, args.buffersize)

	glog.V(1).Infof("End(%s)", time.Since(startTime))

	fmt.Printf("%d\n", linenum)
}
