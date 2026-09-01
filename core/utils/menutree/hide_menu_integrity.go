package menutree

import "github.com/1Panel-dev/1Panel/core/app/dto"

const (
	xpackMenuID    = "11"
	xpackMenuLabel = "Xpack-Menu"
)

func PreserveMissingMenus(menus, fallback []dto.ShowMenu) ([]dto.ShowMenu, bool) {
	updated := cloneMenus(menus)
	changed := preserveMissingMenus(&updated, &updated, fallback)
	return updated, changed
}

func preserveMissingMenus(root, current *[]dto.ShowMenu, fallback []dto.ShowMenu) bool {
	changed := false
	for _, previous := range fallback {
		menuIndex := findMatchingMenu(*current, previous)
		if menuIndex < 0 {
			if containsMenu(*root, previous) {
				continue
			}
			*current = append(*current, cloneMenu(previous))
			changed = true
			continue
		}
		if preserveMissingMenus(root, &(*current)[menuIndex].Children, previous.Children) {
			changed = true
		}
	}
	return changed
}

func ReconcileHideMenuIntegrity(menus, fallback []dto.ShowMenu) ([]dto.ShowMenu, bool) {
	updated := cloneMenus(menus)
	changed := false
	if preferredXApp := findSystemMenu(updated, defaultXAppMenu()); preferredXApp != nil {
		changed = normalizeSystemFields(preferredXApp, defaultXAppMenu())
	}

	var removed bool
	updated, removed = removeSystemMenus(updated, retiredUpageMenuIdentity())
	sanitizedFallback := cloneMenus(fallback)
	if preferredFallbackXApp := findSystemMenu(sanitizedFallback, defaultXAppMenu()); preferredFallbackXApp != nil {
		normalizeSystemFields(preferredFallbackXApp, defaultXAppMenu())
	}
	sanitizedFallback, _ = removeSystemMenus(sanitizedFallback, retiredUpageMenuIdentity())
	changed = changed || removed
	if existingXApp := findSystemMenu(updated, defaultXAppMenu()); existingXApp != nil {
		normalized := normalizeSystemFields(existingXApp, defaultXAppMenu())
		deduplicated, deduplicatedChanged := deduplicateSystemMenu(updated, existingXApp, defaultXAppMenu())
		return deduplicated, changed || normalized || deduplicatedChanged
	}

	parentIndex := findXpackMenu(updated)
	if parentIndex < 0 {
		parent := defaultXpackMenu()
		fallbackParentIndex := findXpackMenu(sanitizedFallback)
		if fallbackParentIndex >= 0 {
			parent = cloneMenu(sanitizedFallback[fallbackParentIndex])
		}
		updated = append(updated, parent)
		parentIndex = len(updated) - 1
		changed = true
	}

	updated, xAppChanged := ensureSystemMenu(
		updated,
		parentIndex,
		sanitizedFallback,
		defaultXAppMenu(),
	)
	changed = changed || xAppChanged

	return updated, changed
}

func defaultXpackMenu() dto.ShowMenu {
	return dto.ShowMenu{
		ID:       xpackMenuID,
		Disabled: false,
		Title:    "xpack.menu",
		IsShow:   true,
		Label:    xpackMenuLabel,
		Sort:     1100,
	}
}

func defaultXAppMenu() dto.ShowMenu {
	return dto.ShowMenu{
		ID:       "118",
		Disabled: false,
		Title:    "xpack.app.app",
		IsShow:   true,
		Label:    "XApp",
		Path:     "/xpack/app",
		Sort:     100,
	}
}

func retiredUpageMenuIdentity() dto.ShowMenu {
	return dto.ShowMenu{
		ID:    "119",
		Label: "Upage",
		Path:  "/xpack/upage",
	}
}

func findXpackMenu(menus []dto.ShowMenu) int {
	for i := range menus {
		if menus[i].ID == xpackMenuID {
			return i
		}
	}
	for i := range menus {
		if menus[i].Label == xpackMenuLabel {
			return i
		}
	}
	return -1
}

func findMenu(menus []dto.ShowMenu, canonical dto.ShowMenu) int {
	for i := range menus {
		if menus[i].ID == canonical.ID {
			return i
		}
	}
	for i := range menus {
		if menus[i].Label == canonical.Label {
			return i
		}
	}
	for i := range menus {
		if menus[i].Path == canonical.Path {
			return i
		}
	}
	return -1
}

func findMatchingMenu(menus []dto.ShowMenu, target dto.ShowMenu) int {
	if target.ID != "" {
		for i := range menus {
			if menus[i].ID == target.ID {
				return i
			}
		}
	}
	if target.Label != "" {
		for i := range menus {
			if menus[i].Label == target.Label {
				return i
			}
		}
	}
	if target.Path != "" {
		for i := range menus {
			if menus[i].Path == target.Path {
				return i
			}
		}
	}
	return -1
}

func containsMenu(menus []dto.ShowMenu, target dto.ShowMenu) bool {
	if findMatchingMenu(menus, target) >= 0 {
		return true
	}
	for i := range menus {
		if containsMenu(menus[i].Children, target) {
			return true
		}
	}
	return false
}

func ensureSystemMenu(menus []dto.ShowMenu, parentIndex int, fallback []dto.ShowMenu, canonical dto.ShowMenu) ([]dto.ShowMenu, bool) {
	selected := findSystemMenu(menus, canonical)
	if selected == nil {
		newItem := canonical
		if fallbackItem := findSystemMenu(fallback, canonical); fallbackItem != nil {
			newItem = cloneMenu(*fallbackItem)
			normalizeSystemFields(&newItem, canonical)
		}
		menus[parentIndex].Children = append(menus[parentIndex].Children, newItem)
		return menus, true
	}

	changed := normalizeSystemFields(selected, canonical)
	deduplicated, deduplicatedChanged := deduplicateSystemMenu(menus, selected, canonical)
	return deduplicated, changed || deduplicatedChanged
}

func findSystemMenu(menus []dto.ShowMenu, canonical dto.ShowMenu) *dto.ShowMenu {
	matchers := []func(dto.ShowMenu) bool{
		func(menu dto.ShowMenu) bool { return menu.ID == canonical.ID },
		func(menu dto.ShowMenu) bool { return menu.Label == canonical.Label },
		func(menu dto.ShowMenu) bool { return menu.Path == canonical.Path },
	}
	for _, matches := range matchers {
		if menu := findMenuBy(menus, matches); menu != nil {
			return menu
		}
	}
	return nil
}

func findMenuBy(menus []dto.ShowMenu, matches func(dto.ShowMenu) bool) *dto.ShowMenu {
	for i := range menus {
		if matches(menus[i]) {
			return &menus[i]
		}
		if menu := findMenuBy(menus[i].Children, matches); menu != nil {
			return menu
		}
	}
	return nil
}

func deduplicateSystemMenu(menus []dto.ShowMenu, selected *dto.ShowMenu, canonical dto.ShowMenu) ([]dto.ShowMenu, bool) {
	deduplicated := make([]dto.ShowMenu, 0, len(menus))
	changed := false
	for i := range menus {
		menu := &menus[i]
		if menu != selected && matchesSystemIdentity(*menu, canonical) {
			changed = true
			continue
		}
		children, childChanged := deduplicateSystemMenu(menu.Children, selected, canonical)
		if childChanged {
			menu.Children = children
			changed = true
		}
		deduplicated = append(deduplicated, *menu)
	}
	return deduplicated, changed
}

func removeSystemMenus(menus []dto.ShowMenu, identity dto.ShowMenu) ([]dto.ShowMenu, bool) {
	if menus == nil {
		return nil, false
	}
	filtered := make([]dto.ShowMenu, 0, len(menus))
	changed := false
	for i := range menus {
		menu := &menus[i]
		children, childChanged := removeSystemMenus(menu.Children, identity)
		if matchesSystemIdentity(*menu, identity) {
			filtered = append(filtered, children...)
			changed = true
			continue
		}
		if childChanged {
			menu.Children = children
			changed = true
		}
		filtered = append(filtered, *menu)
	}
	return filtered, changed
}

func matchesSystemIdentity(menu, canonical dto.ShowMenu) bool {
	return menu.ID == canonical.ID || menu.Label == canonical.Label || menu.Path == canonical.Path
}

func normalizeSystemFields(menu *dto.ShowMenu, canonical dto.ShowMenu) bool {
	changed := menu.ID != canonical.ID ||
		menu.Label != canonical.Label ||
		menu.Disabled != canonical.Disabled ||
		menu.Title != canonical.Title ||
		menu.Path != canonical.Path

	menu.ID = canonical.ID
	menu.Label = canonical.Label
	menu.Disabled = canonical.Disabled
	menu.Title = canonical.Title
	menu.Path = canonical.Path
	return changed
}

func cloneMenus(menus []dto.ShowMenu) []dto.ShowMenu {
	if menus == nil {
		return nil
	}
	cloned := make([]dto.ShowMenu, len(menus))
	for i := range menus {
		cloned[i] = cloneMenu(menus[i])
	}
	return cloned
}

func cloneMenu(menu dto.ShowMenu) dto.ShowMenu {
	menu.Children = cloneMenus(menu.Children)
	return menu
}
