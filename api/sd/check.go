// Package api provides APi for chestnut.
package sd

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)


// @Summary Shows OK as the ping-pong result
// @Description Shows OK as the ping-pong result
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain "OK"
// @Router /sd/health [get]
func HealthCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		output := "OK"
		return c.JSON(http.StatusOK, output)
	}
}

// @Summary Checks the cpu usage
// @Description Checks the cpu usage
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain "CRITICAL - Load average: 1.78, 1.99, 2.02 | Cores: 2"
// @Router /sd/cpu [get]
func CPUCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		cores, _ := cpu.Counts(false)
		a, _ := load.Avg()

		l1 := a.Load1
		l5 := a.Load5
		l15 := a.Load15

		status := http.StatusOK
		text := "OK"

		if l5 >= float64(cores-1) {
			status = http.StatusInternalServerError
			text = "CRITICAL"
		} else if l5 >= float64(cores-2) {
			status = http.StatusTooManyRequests
			text = "WARNING"
		}

		message := fmt.Sprintf("%s - Load average: %.2f, %.2f, %.2f | Cores: %d", text, l1, l5, l15, cores)
		return c.JSON(status, message)
	}
}


// @Summary Checks the disk usage
// @Description Checks the disk usage
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain "OK - Free space: 17233MB (16GB) / 51200MB (50GB) | Used: 33%"
// @Router /sd/disk [get]
func DiskCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		u, _ := disk.Usage("/")

		usedMB := int(u.Used) / MB
		usedGB := int(u.Used) / GB
		totalMB := int(u.Total) / MB
		totalGB := int(u.Total) / GB
		usedPercent := int(u.UsedPercent)

		status := http.StatusOK
		text := "OK"

		if usedPercent >= 95 {
			status = http.StatusOK
			text = "CRITICAL"
		} else if usedPercent >= 90 {
			status = http.StatusTooManyRequests
			text = "WARNING"
		}

		message := fmt.Sprintf("%s - Free space: %dMB (%dGB) / %dMB (%dGB) | Used: %d%%", text, usedMB, usedGB, totalMB, totalGB, usedPercent)
		return c.JSON(status, message)
	}
}




// @Summary Checks the ram usage
// @Description Checks the ram usage
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain "OK - Free space: 402MB (0GB) / 8192MB (8GB) | Used: 4%"
// @Router /sd/ram [get]
func RAMCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		u, _ := mem.VirtualMemory()

		usedMB := int(u.Used) / MB
		usedGB := int(u.Used) / GB
		totalMB := int(u.Total) / MB
		totalGB := int(u.Total) / GB
		usedPercent := int(u.UsedPercent)

		status := http.StatusOK
		text := "OK"

		if usedPercent >= 95 {
			status = http.StatusInternalServerError
			text = "CRITICAL"
		} else if usedPercent >= 90 {
			status = http.StatusTooManyRequests
			text = "WARNING"
		}

		message := fmt.Sprintf("%s - Free space: %dMB (%dGB) / %dMB (%dGB) | Used: %d%%", text, usedMB, usedGB, totalMB, totalGB, usedPercent)
		return c.JSON(status, message)
	}
}


// @Summary Checks the host message
// @Description Checks the host message
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain hostname : LAPTOP | Platform : Microsoft Windows 10 Home China | KernelArch : x86_64
// @Router /sd/host [get]
func HostCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		status := http.StatusOK

		hInfo, _ := host.Info()
		message := fmt.Sprintf("hostname : %s | Platform : %s | KernelArch : %s",hInfo.Hostname,hInfo.Platform,hInfo.KernelArch)
		return c.JSON(status, message)
	}
}

// @Summary Checks the net message
// @Description Checks the net message
// @Tags sd
// @Accept  json
// @Produce  json
// @Success 200 {string} plain WLAN BytesSent : 9328315 | BytesRecv : 87317530 | packetsSent : 65301 |  packageRecv : 81460
// @Router /sd/net [get]
func NetCheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		infoN, _ := net.IOCounters(true)
		status := http.StatusOK

		for _, v := range infoN {
			message := fmt.Sprintf("%v BytesSent : %v | BytesRecv : %v | packetsSent : %v |  packageRecv : %v\n", v.Name, v.BytesSent, v.BytesRecv,v.PacketsSent,v.PacketsRecv)
			return c.JSON(status, "\n" + message)
		}
		return c.JSON(status, "")
	}
}
