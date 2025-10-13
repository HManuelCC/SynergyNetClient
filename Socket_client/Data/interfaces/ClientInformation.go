package interfaces

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type ClientHardwareResourcesStatistics struct {
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	DiskBusy    float64
}

type ClientInformation struct {
	ClientName string                            `json:"client_name"`
	Latency    float64                           `json:"latency"`
	Resources  ClientHardwareResourcesStatistics `json:"resources"`
}

func (c *ClientHardwareResourcesStatistics) GetSystemStats() error {
	// CPU (promedio en 1 segundo)
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}
	if len(cpuPercent) > 0 {
		c.CPUUsage = cpuPercent[0]
	}

	// Memoria
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	c.MemoryUsage = vmStat.UsedPercent

	diskStat, err := disk.Usage("/")
	if err != nil {
		return err
	}
	c.DiskUsage = diskStat.UsedPercent

	ioStart, err := disk.IOCounters()
	if err != nil {
		return err
	}

	// Disco
	time.Sleep(1 * time.Second)

	ioEnd, err := disk.IOCounters()
	if err != nil {
		return err
	}

	// calcular diferencias (para el primer disco encontrado)
	for name, start := range ioStart {
		end := ioEnd[name]

		readDelta := end.ReadBytes - start.ReadBytes
		writeDelta := end.WriteBytes - start.WriteBytes

		// aquí solo como ejemplo: ocupación relativa (simplificada)
		totalIO := readDelta + writeDelta
		if totalIO > 0 {
			// asumimos que un disco puede mover ~100MB/s (depende del HW real)
			maxThroughput := float64(100 * 1024 * 1024) // 100 MB/s
			c.DiskBusy = (float64(totalIO) / maxThroughput) * 100
			if c.DiskBusy > 100 {
				c.DiskBusy = 100
			}
		}

		break // solo un disco para simplificar
	}
	return nil
}
