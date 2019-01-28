package nodejs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	// nsenter ensures something in the 2-step process of spawning a container
	_ "github.com/opencontainers/runc/libcontainer/nsenter"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/specconv"
)

// Nodejs represents a nodejs process
// one should write javascript code to it
// and read results from it
type Nodejs struct {
	*io.PipeWriter
	*io.PipeReader

	Stderr io.Reader
}

// this init takes over the entire app when run with the second argument "init"
// the documentation mentions something about a 2 step process involved in starting
// a container
func init() {

	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()

		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			panic(err)
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

// NewNodejs returns a Nodejs which can be written javascript and read results from
// it does a 5 second timeout, which should be more then enough as no IO can happen
// (ish, filesystems are readonly and there are no network interfaces setup)
func NewNodejs() (*Nodejs, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// use example spec - it seems restrictive enough
	spec := specconv.Example()

	// change root directory (and mout readonly)
	spec.Root.Readonly = true
	spec.Root.Path = path.Join(cwd, "rootfs")
	log.Printf("launching with rootfs %s", spec.Root.Path)

	// create a config from the default spec
	config, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		CgroupName: "dbgauge",
		Spec:       spec,
	})

	if err != nil {
		return nil, err
	}

	// create a container factory
	factory, err := libcontainer.New("/tmp/DanskeBankGauge-libcontainer", libcontainer.Cgroupfs, libcontainer.InitArgs(os.Args[0], "init"))
	if err != nil {
		return nil, err
	}

	// create the nodejs container, name it nodejs-container
	container, err := factory.Create("nodejs-container", config)
	if err != nil {
		return nil, err
	}

	// Some pipes for stdio comms
	stdoutReader, stdoutWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()
	errorBuffer := &bytes.Buffer{}

	// we are ready to setup our Nodejs struct
	n := Nodejs{PipeReader: stdoutReader, PipeWriter: stdinWriter, Stderr: errorBuffer}

	process := &libcontainer.Process{
		Args:   []string{"node"},
		Env:    []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		Stdin:  stdinReader,
		Stdout: stdoutWriter,
		Stderr: errorBuffer,
		Init:   true,
	}

	err = container.Run(process)
	if err != nil {
		container.Destroy()
		return nil, fmt.Errorf("unable to run process: %s", err)
	}

	// Setup timeout
	go func() {
		time.Sleep(time.Second * 5)
		stdinWriter.Close()
		process.Signal(os.Kill)
	}()

	// ensure the process is destroyed when finished
	go func() {
		// wait for the process to finish.
		_, err := process.Wait()
		if err != nil {
			stdoutWriter.CloseWithError(err)
		}

		// destroy the container.
		container.Destroy()
		stdinWriter.Close()
		stdoutWriter.Close()
	}()

	return &n, nil
}

// Close closes in stdin pipe
func (n *Nodejs) Close() {
	n.PipeWriter.Close()
}

// Eval evaluates code read from io.Reader
func Eval(code io.Reader) ([]byte, error) {

	n, err := NewNodejs()
	if err != nil {
		return nil, fmt.Errorf("unable to initiate nodejs: %s", err)
	}
	// tmpFile, _ := ioutil.TempFile("/tmp", "javascræpt")

	// io.Copy(tmpFile, code)
	// log.Printf("%s written", tmpFile.Name())
	// n.Write([]byte("console.log('her er javascript', true, 3245); process.exit(0);"))
	io.Copy(n, code)
	n.Close()

	result, err := ioutil.ReadAll(n)
	if err != nil {
		stderr, _ := ioutil.ReadAll(n.Stderr)
		return nil, fmt.Errorf("error when evaluating: %s\n %s", err, stderr)
	}

	return result, nil
}
