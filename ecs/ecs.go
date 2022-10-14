package ecs

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/query"
)

// ECS represents an entity-component-system.
type ECS struct {
	// World is the underlying world of the ECS.
	World donburi.World
	// Time manages the time of the world.
	Time *Time
	// ScriptSystem manages the scripts of the world.
	ScriptSystem *ScriptSystem

	*innerECS
}

// SystemOpts represents options for systems.
type SystemOpts struct {
	// Image is the image to draw the system.
	Image *ebiten.Image
	// Priority is the priority of the system.
	Priority int
}

type innerECS struct {
	updaters []*updater
	drawers  []*drawer
}

type updater struct {
	Updater  Updater
	Priority int
}

type drawer struct {
	Drawer   Drawer
	Priority int
	Image    *ebiten.Image
}

// NewECS creates a new ECS with the specified world.
func NewECS(w donburi.World) *ECS {
	ecs := &ECS{
		World: w,
		Time:  NewTime(),
		innerECS: &innerECS{
			updaters: []*updater{},
			drawers:  []*drawer{},
		},
	}

	ecs.ScriptSystem = NewScriptSystem()
	ecs.AddSystem(ecs.ScriptSystem, &SystemOpts{})

	return ecs
}

// AddSystems adds new systems either Updater or Drawer
func (ecs *ECS) AddSystems(systems ...interface{}) {
	for _, system := range systems {
		ecs.AddSystem(system, nil)
	}
}

// AddSystem adds new system either Updater or Drawer
func (ecs *ECS) AddSystem(system interface{}, opts *SystemOpts) {
	if opts == nil {
		opts = &SystemOpts{}
	}
	flag := false
	if system, ok := system.(Updater); ok {
		ecs.addUpdater(&updater{
			Updater:  system,
			Priority: opts.Priority,
		})
		flag = true
	}
	if system, ok := system.(Drawer); ok {
		ecs.addDrawer(&drawer{
			Drawer:   system,
			Priority: opts.Priority,
			Image:    opts.Image,
		})
		flag = true
	}
	if !flag {
		panic("ECS system should be either Updater or Drawer at least.")
	}
}

// AddScript adds a script to the specified entity.
// the argument `script` must implement EntryUpdater or/and EntryDrawer interface.
func (ecs *ECS) AddScript(q *query.Query, script interface{}, opts *ScriptOpts) {
	ecs.ScriptSystem.AddScript(q, script, opts)
}

// Update calls Updater's Update() methods.
func (ecs *ECS) Update() {
	ecs.Time.Update()
	for _, u := range ecs.updaters {
		u.Updater.Update(ecs)
	}
}

// AddUpdaterWithPriority adds an Updater to the ECS with the specified priority.
// Higher priority is executed first.
func (ecs *ECS) addUpdater(entry *updater) {
	ecs.updaters = append(ecs.updaters, entry)
	sortUpdaterEntries(ecs.updaters)
}

// AddDrawer adds a Drawer to the ECS.
func (ecs *ECS) addDrawer(entry *drawer) {
	ecs.drawers = append(ecs.drawers, entry)
	sortDrawerEntries(ecs.drawers)
}

// Draw calls Drawer's Draw() methods.
func (ecs *ECS) Draw(screen *ebiten.Image) {
	for _, d := range ecs.drawers {
		if d.Image != nil {
			d.Drawer.Draw(ecs, d.Image)
			continue
		}
		d.Drawer.Draw(ecs, screen)
	}
}

func sortUpdaterEntries(entries []*updater) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Priority > entries[j].Priority
	})
}

func sortDrawerEntries(entries []*drawer) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Priority > entries[j].Priority
	})
}
