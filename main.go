package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateChooseDisk state = iota
	stateEnterMountPoint
	stateEnterPermissions
	stateFormatAndMount
	stateDone
)

type DiskItem struct {
	name string
}

func (d DiskItem) Title() string       { return d.name }
func (d DiskItem) Description() string { return "Disk available for formatting" }
func (d DiskItem) FilterValue() string { return d.name }

type tickMsg struct{}

type loadingModel struct {
	tickIndex int
	ticks     []string
}

func newLoadingModel() loadingModel {
	return loadingModel{
		tickIndex: 0,
		ticks:     []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (lm loadingModel) next() loadingModel {
	return loadingModel{
		tickIndex: (lm.tickIndex + 1) % len(lm.ticks),
		ticks:     lm.ticks,
	}
}

func (lm loadingModel) currentTick() string {
	return lm.ticks[lm.tickIndex]
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type model struct {
	state        state
	disks        []DiskItem
	selectedDisk string
	mountPoint   string
	permissions  string
	list         list.Model
	textInput    textinput.Model
	loading      loadingModel
	err          error
}

func initialModel() model {
	disks, err := getAvailableDisks()
	if err != nil {
		log.Fatalf("Error fetching disks: %v\n", err)
	}

	items := make([]list.Item, len(disks))
	for i, disk := range disks {
		items[i] = DiskItem{name: disk.name}
	}

	const defaultWidth = 20
	const defaultHeight = 10

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	l.Title = "Select a disk"

	ti := textinput.New()
	ti.Placeholder = "Enter mount point (e.g., /mnt/data)"
	ti.Focus()

	return model{
		state:     stateChooseDisk,
		disks:     disks,
		list:      l,
		textInput: ti,
		loading:   newLoadingModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateChooseDisk:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				selected, ok := m.list.SelectedItem().(DiskItem)
				if ok {
					m.selectedDisk = selected.name
					m.state = stateEnterMountPoint
				}
				m.textInput.SetValue("")
				return m, nil
			case "q":
				return m, tea.Quit
			}
		}
		m.list, _ = m.list.Update(msg)

	case stateEnterMountPoint:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.mountPoint = m.textInput.Value()
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter permissions (e.g., 0755)"
				m.state = stateEnterPermissions
				return m, nil
			case "q":
				return m, tea.Quit
			}
		}
		m.textInput, _ = m.textInput.Update(msg)

	case stateEnterPermissions:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.permissions = m.textInput.Value()
				m.state = stateFormatAndMount
				return m, tea.Batch(m.formatAndMount(), tickCmd())
			case "q":
				return m, tea.Quit
			}
		}
		m.textInput, _ = m.textInput.Update(msg)

	case stateFormatAndMount:
		switch msg := msg.(type) {
		case tickMsg:
			m.loading = m.loading.next()
			return m, tickCmd()
		case string:
			m.state = stateDone
			m.err = nil
			return m, nil
		case error:
			m.state = stateDone
			m.err = msg
			return m, nil
		}

	case stateDone:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateChooseDisk:
		return m.list.View()
	case stateEnterMountPoint:
		return fmt.Sprintf("Enter mount point: %s\n", m.textInput.View())
	case stateEnterPermissions:
		return fmt.Sprintf("Enter permissions (e.g., 0755): %s\n", m.textInput.View())
	case stateFormatAndMount:
		return fmt.Sprintf("Formatting and mounting the disk...\n%s\nPress q to quit.", m.loading.currentTick())
	case stateDone:
		if m.err != nil {
			return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
		}
		return "Disk successfully formatted and mounted!\nPress q to quit."
	}
	return ""
}

func (m model) formatAndMount() tea.Cmd {
	return func() tea.Msg {
		perm, err := strconv.ParseUint(m.permissions, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid permissions: %w", err)
		}

		// Format the disk
		if err := exec.Command("sudo", "mkfs.ext4", m.selectedDisk).Run(); err != nil {
			return fmt.Errorf("failed to format disk: %w", err)
		}

		// Create mount point
		if err := os.MkdirAll(m.mountPoint, os.FileMode(perm)); err != nil {
			return fmt.Errorf("failed to create mount point: %w", err)
		}

		// Mount the disk
		if err := exec.Command("sudo", "mount", m.selectedDisk, m.mountPoint).Run(); err != nil {
			return fmt.Errorf("failed to mount disk: %w", err)
		}

		// Get UUID and update /etc/fstab
		output, err := exec.Command("sudo", "blkid", "-s", "UUID", "-o", "value", m.selectedDisk).Output()
		if err != nil {
			return fmt.Errorf("failed to get disk UUID: %w", err)
		}
		uuid := strings.TrimSpace(string(output))
		fstabEntry := fmt.Sprintf("UUID=%s %s ext4 defaults,nofail 0 2\n", uuid, m.mountPoint)
		fstabFile, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to update /etc/fstab: %w", err)
		}
		defer fstabFile.Close()
		if _, err := fstabFile.WriteString(fstabEntry); err != nil {
			return fmt.Errorf("failed to write to /etc/fstab: %w", err)
		}

		// Test mount
		if err := exec.Command("sudo", "mount", "-a").Run(); err != nil {
			return fmt.Errorf("failed to test mount: %w", err)
		}

		return "Disk successfully formatted and mounted!"
	}
}

func getAvailableDisks() ([]DiskItem, error) {
	cmd := exec.Command("lsblk", "-dn", "-o", "NAME")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(out.String(), "\n")
	var disks []DiskItem
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			disks = append(disks, DiskItem{name: "/dev/" + line})
		}
	}
	return disks, nil
}

func main() {
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v\n", err)
	}

	if m, ok := finalModel.(model); ok {
		if m.err != nil {
			fmt.Printf("An error occurred: %v\n", m.err)
		} else {
			fmt.Println("Program completed successfully!")
		}
	}
}
