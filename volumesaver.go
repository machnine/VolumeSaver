package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Volume struct {
	ContainerPath string `json:"container_path"`
}

type Config struct {
	ContainerName string   `json:"container_name"`
	BackupDir     string   `json:"backup_dir"`
	TempDir       string   `json:"temp_dir"`
	Volumes       []Volume `json:"volumes"`
}

type DockerVolumeBackup struct {
	Config Config
}

func NewDockerVolumeBackup(config Config) *DockerVolumeBackup {
	return &DockerVolumeBackup{Config: config}
}

func (d *DockerVolumeBackup) createTempDir() error {
	return os.MkdirAll(d.Config.TempDir, os.ModePerm)
}

func (d *DockerVolumeBackup) cleanupTempDir() error {
	return os.RemoveAll(d.Config.TempDir)
}

func (d *DockerVolumeBackup) getTimestamp() string {
	return time.Now().Format("20060102150405")
}

func (d *DockerVolumeBackup) copyDataFromContainer(containerPath string) error {
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", d.Config.ContainerName, containerPath), d.Config.TempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *DockerVolumeBackup) compressBackup(volume Volume) error {
	timestamp := d.getTimestamp()
	zipFile := filepath.Join(d.Config.BackupDir, fmt.Sprintf("%s_%s.zip", filepath.Base(volume.ContainerPath), timestamp))

	file, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	err = filepath.Walk(d.Config.TempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(d.Config.TempDir, path)
		if err != nil {
			return err
		}

		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(zipFileWriter, srcFile)
		return err
	})

	return err
}

func (d *DockerVolumeBackup) Backup() error {
	for _, volume := range d.Config.Volumes {
		fmt.Printf("Backing up volume: %s\n", volume.ContainerPath)

		if err := d.createTempDir(); err != nil {
			return err
		}

		if err := d.copyDataFromContainer(volume.ContainerPath); err != nil {
			return err
		}

		if err := d.compressBackup(volume); err != nil {
			return err
		}

		if err := d.cleanupTempDir(); err != nil {
			return err
		}
	}
	return nil
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func main() {
	configFile := flag.String("c", "config.json", "Configuration file")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	backup := NewDockerVolumeBackup(config)
	if err := backup.Backup(); err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Backup succeeded.")
	}
}
