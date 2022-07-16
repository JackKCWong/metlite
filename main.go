package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/procfs"
)

type cpuStat struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	// Boot time in seconds since the Epoch.
	BootTime uint64
	// Summed up cpu statistics.
	Stat procfs.CPUStat `gorm:"embedded"`
	// Per-CPU statistics.
	CPUs []cpuCoreStat
	// Number of times interrupts were handled, which contains numbered and unnumbered IRQs.
	IRQTotal uint64
	// Number of times a context switch happened.
	ContextSwitches uint64
	// Number of times a process was created.
	ProcessCreated uint64
	// Number of processes currently running.
	ProcessesRunning uint64
	// Number of processes currently blocked (waiting for IO).
	ProcessesBlocked uint64
	// Number of times a softirq was scheduled.
	SoftIRQTotal uint64
}

type cpuCoreStat struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	NO        uint
	Stat      procfs.CPUStat `gorm:"embedded"`
	CpuStatID uint
}

type memStat struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	Stat      procfs.Meminfo `gorm:"embedded"`
}

func main() {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		log.Panicf("failed to open /proc: %q", err)
	}

	db, err := openDB("test.db")
	if err != nil {
		log.Panicf("failed to open test.db: %q", err)
	}

	db.AutoMigrate(&cpuStat{}, &cpuCoreStat{}, &memStat{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigCh:
			log.Printf("exit")
			os.Exit(0)
		default:
			cpu, err := getCpuStat(fs)
			if err != nil {
				log.Printf("cpu stat err: %q", err)
			}

			result := db.Create(&cpu)
			if result.Error != nil {
				log.Printf("cpu stat save err: %q", result.Error)
			}

			var mem memStat
			meminfo, err := fs.Meminfo()
			if err != nil {
				log.Printf("mem stat err: %q", err)
			}

			mem.Stat = meminfo

			result = db.Create(&mem)
			if result.Error != nil {
				log.Printf("mem stat save err: %q", result.Error)
			}

			log.Printf("%+v", cpu)
			log.Printf("%+v", mem)
			time.Sleep(1 * time.Second)
		}
	}
}

func getCpuStat(fs procfs.FS) (cpuStat, error) {
	var rec cpuStat
	stat, err := fs.Stat()
	if err != nil {
		return cpuStat{}, err
	}

	rec.BootTime = stat.BootTime
	rec.Stat = stat.CPUTotal
	rec.ContextSwitches = stat.ContextSwitches
	rec.ProcessCreated = stat.ProcessCreated
	rec.ProcessesBlocked = stat.ProcessesBlocked
	rec.ProcessesRunning = stat.ProcessesRunning
	rec.IRQTotal = stat.IRQTotal
	rec.SoftIRQTotal = stat.SoftIRQTotal

	cpus := make([]cpuCoreStat, 0, len(stat.CPU))
	for i := range stat.CPU {
		var cpu cpuCoreStat
		cpu.Stat = stat.CPU[i]
		cpu.NO = uint(i)
		cpus = append(cpus, cpu)
	}

	rec.CPUs = cpus

	return rec, nil
}
