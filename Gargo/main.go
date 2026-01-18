package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"io"
	"github.com/pkg/sftp"
)

func SyncDirectory(rc *RemoteClient, localdir, remotedir string, recursive bool) error {
	sc, err := sftp.NewClient(rc.Client)
	if err != nil {
		return err
	}
	defer sc.Close()

	return filepath.Walk(localdir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }

		relpath, _ := filepath.Rel(localdir, path)
		if relpath == "." { return nil }
		
		remotepath := filepath.ToSlash(filepath.Join(remotedir, relpath))

		if info.IsDir() {
			if !recursive { return filepath.SkipDir }
			return sc.MkdirAll(remotepath)
		}

		localfile, err := os.Open(path)
		if err != nil { return err }
		defer localfile.Close()

		sc.MkdirAll(filepath.Dir(remotepath))
		remotefile, err := sc.Create(remotepath)
		if err != nil { return err }
		defer remotefile.Close()

		_, err = io.Copy(remotefile, localfile)
		return err
	})
}

func main() {
	cfg, err := Parse("config.ini")
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	p := cfg["project"]
	r := cfg["remote"]
	b := cfg["build"]

	port := r["port"]
	if port == "" {
		port = "22"
	}
	password := r["password"]
	if password == "" {
		log.Fatal("Error: 'password' field is missing in config.ini")
	}

	fulladdr := fmt.Sprintf("%s:%s", r["host"], port)

	fmt.Printf("Connecting to %s@%s...\n", r["user"], fulladdr)
	client, err := Connect(r["user"], fulladdr, password)
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer client.Close()

	inputsetting := p["inputdir"]
	inputdir := "mycodes"
	recursive := false

	if strings.Contains(inputsetting, ",") {
		parts := strings.Split(inputsetting, ",")
		inputdir = strings.TrimSpace(parts[0])
		recursive = strings.TrimSpace(parts[1]) == "true"
	} else {
		inputdir = inputsetting
	}

	fmt.Printf("Scanning Source: %s (Recursive: %v)\n", inputdir, recursive)

	initcommands := []string{"mkdir -p " + p["directory"]}
	initcommands = append(initcommands,
		"cd "+p["directory"],
		fmt.Sprintf("go mod init %s || true", p["name"]),
		"go mod tidy",
	)

	fmt.Println("--- [UPLOAD & INIT PHASE] ---")
	initstart := time.Now()
	err = SyncDirectory(client, inputdir, p["directory"], recursive)
	if err != nil {
		log.Fatalf("Upload phase failed: %v", err)
	}

	buildflags := ""
	if p["small"] == "true" {
		buildflags = "-trimpath -ldflags=\"-s -w\""
	}
	
	gogc := b["GOGC"]
	if gogc == "" {
		gogc = "20"
	}
	
	out, err := client.RunCommand("uname -s 2>/dev/null || cmd /c ver")
	if err != nil {
		log.Fatalf("Unable to detect remote OS: %v", err)
	}

	u := strings.ToLower(out)
	remoteos := "unknown"

	if strings.Contains(u, "linux") {
		remoteos = "linux"
	} else if strings.Contains(u, "windows") {
		remoteos = "windows"
	}

	
	expectedremote := r["expectedremote"]
	if expectedremote == "" {
		expectedremote = "linux"
	}
	expectedremote = strings.ToLower(expectedremote)

	if expectedremote != "any" && remoteos != expectedremote {
		log.Fatalf(
			"Remote OS mismatch: expected %s, got %s",
			expectedremote,
			remoteos,
		)
	}
	
	cgo := b["CGO_ENABLED"]
	if cgo == "" {
		cgo = "0"
	}

	cores := b["core"]
	if cores == "" {
		cores = "1"
	}

	buildcmd := fmt.Sprintf("cd %s && GOGC=%s CGO_ENABLED=%s GOOS=%s GOARCH=%s go build -p %s %s -o %s",
		p["directory"],
		gogc,
		cgo,
		b["os"],
		b["arch"],
		cores,
		buildflags,
		p["name"],
	)

	fmt.Printf(">> Init & Upload took: %v\n", time.Since(initstart))

	fmt.Printf("\n--- [BUILD PHASE] (%s/%s) ---\n", b["os"], b["arch"])
	buildstart := time.Now()

	if b["before"] != "" {
		fmt.Println("--- [PRE-BUILD HOOKS] ---")
		beforecmds := strings.Split(b["before"], ",")
		for _, cmd := range beforecmds {
			fullcmd := fmt.Sprintf("cd %s && %s", p["directory"], strings.TrimSpace(cmd))
			client.RunScript(fullcmd)
		}
	}

	if err := client.RunScript(buildcmd); err != nil {
		log.Fatalf("Build phase failed: %v", err)
	}

	fmt.Printf(">> Build took: %v\n", time.Since(buildstart))

	if b["after"] != "" {
		fmt.Println("\n--- [POST-BUILD HOOKS] ---")
		aftercmds := strings.Split(b["after"], ",")
		for _, cmd := range aftercmds {
			fullcmd := fmt.Sprintf("cd %s && %s", p["directory"], strings.TrimSpace(cmd))
			client.RunScript(fullcmd)
		}
	}

	fmt.Println("\n--- [DOWNLOAD PHASE] ---")

	remotefilename := p["name"]
	remotefilepath := p["directory"] + "/" + remotefilename
	localfilepath := "./" + remotefilename

	fmt.Printf("Downloading binary: %s -> %s\n", remotefilepath, localfilepath)
	err = client.Download(remotefilepath, localfilepath)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Printf("\n[SUCCESS] %s compiled and downloaded successfully!\n", p["name"])
}