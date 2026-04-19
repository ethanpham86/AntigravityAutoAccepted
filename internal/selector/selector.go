// Package selector provides interactive screen region selection.
package selector

import (
	"bytes"
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const psScript = `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$rect = [System.Drawing.Rectangle]::Empty
$form = New-Object System.Windows.Forms.Form
$form.FormBorderStyle = 'None'
$form.StartPosition = 'Manual'
$form.Location = [System.Windows.Forms.SystemInformation]::VirtualScreen.Location
$form.Size = [System.Windows.Forms.SystemInformation]::VirtualScreen.Size
$form.TopMost = $true
$form.BackColor = 'Black'
$form.Opacity = 0.3
$form.Cursor = [System.Windows.Forms.Cursors]::Cross

$script:isDragging = $false
$script:startPoint = [System.Drawing.Point]::Empty

$form.Add_MouseDown({
    if ($_.Button -eq [System.Windows.Forms.MouseButtons]::Left) {
        $script:isDragging = $true
        $script:startPoint = $_.Location
    }
})

$form.Add_MouseMove({
    if ($script:isDragging) {
        $p1 = $script:startPoint
        $p2 = $_.Location
        $x = [math]::Min($p1.X, $p2.X)
        $y = [math]::Min($p1.Y, $p2.Y)
        $w = [math]::Abs($p1.X - $p2.X)
        $h = [math]::Abs($p1.Y - $p2.Y)
        $script:rect = New-Object System.Drawing.Rectangle($x, $y, $w, $h)
        $form.Invalidate()
    }
})

$form.Add_MouseUp({
    if ($_.Button -eq [System.Windows.Forms.MouseButtons]::Left) {
        $script:isDragging = $false
        $form.DialogResult = [System.Windows.Forms.DialogResult]::OK
        $form.Close()
    }
})

$form.Add_Paint({
    if ($script:rect.Width -gt 0 -and $script:rect.Height -gt 0) {
        $pen = New-Object System.Drawing.Pen([System.Drawing.Color]::Red, 2)
        $_.Graphics.DrawRectangle($pen, $script:rect)
        $brush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::FromArgb(50, 255, 0, 0))
        $_.Graphics.FillRectangle($brush, $script:rect)
    }
})

$form.ShowDialog() | Out-Null
Write-Output "$($script:rect.X),$($script:rect.Y),$($script:rect.Width),$($script:rect.Height)"
`

// SelectRegion interactively prompts the user to select a screen region using a visual overlay.
func SelectRegion() (image.Rectangle, error) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        SCREEN REGION SELECTION                  ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  >>> Kéo thả chuột để khoanh vùng lựa chọn      ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	
	err := cmd.Run()
	if err != nil {
		return image.Rectangle{}, fmt.Errorf("failed to run selection script: %w", err)
	}

	result := strings.TrimSpace(out.String())
	parts := strings.Split(result, ",")
	if len(parts) != 4 {
		return image.Rectangle{}, fmt.Errorf("không có vùng nào được chọn, huỷ bỏ thao tác")
	}

	x, err1 := strconv.Atoi(parts[0])
	y, err2 := strconv.Atoi(parts[1])
	w, err3 := strconv.Atoi(parts[2])
	h, err4 := strconv.Atoi(parts[3])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return image.Rectangle{}, fmt.Errorf("lỗi phân tích toạ độ click chuột: %s", result)
	}

	rect := image.Rect(x, y, x+w, y+h)
	if rect.Dx() < 10 || rect.Dy() < 10 {
		return image.Rectangle{}, fmt.Errorf("vùng chọn quá nhỏ: %dx%d", rect.Dx(), rect.Dy())
	}

	fmt.Printf("    ✓ Chọn thành công vùng: %dx%d pixel tại Toạ độ (%d,%d)\n\n", rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y)
	return rect, nil
}
