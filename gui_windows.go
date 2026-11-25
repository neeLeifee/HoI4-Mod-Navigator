package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var uCreate = user.NewProc("CreateWindowExW")
var send = user.NewProc("SendMessageW")
var setText = user.NewProc("SetWindowTextW")
var getText = user.NewProc("GetWindowTextW")
var getLen = user.NewProc("GetWindowTextLengthW")
var def = user.NewProc("DefWindowProcW")
var winMessage = user.NewProc("MessageBoxW")
var instance uintptr
var hwnd uintptr
var font, titleFont, sectionFont uintptr
var labelColors = map[uintptr]uintptr{}
var lastControl uintptr
var gdi = syscall.NewLazyDLL("gdi32.dll")
var panelBrush uintptr
var highContrast bool
var panelColor uintptr = 0xffffff
var ctl = map[int]uintptr{}
var config Config

const (
	idGame       = 101
	idBrowseGame = 102
	idList       = 103
	idName       = 104
	idPath       = 105
	idBrowseMod  = 106
	idAdd        = 107
	idUpdate     = 108
	idDelete     = 109
	idOpen       = 110
	idCopy       = 111
	idReverse    = 112
	idCompare    = 113
	idCreate     = 114
	idFilter     = 115
	idApply      = 116
	idRemove     = 117
	idClose      = 118
	idStatus     = 119
	idLegacy     = 120
)

func message(owner uintptr, s string, flags uintptr) uintptr {
	r, _, _ := winMessage.Call(owner, ptr(wide(s)), ptr(wide(appName)), flags|0x10000)
	return r
}
func ask(s string) bool { return message(hwnd, s, 0x04|0x20|0x100) == 6 }
func textOf(id int) string {
	h := ctl[id]
	n, _, _ := getLen.Call(h)
	b := make([]uint16, n+1)
	getText.Call(h, uintptr(unsafe.Pointer(&b[0])), n+1)
	return syscall.UTF16ToString(b)
}
func set(id int, s string) { setText.Call(ctl[id], ptr(wide(s))) }
func status(s string)      { set(idStatus, s) }
func selected() int        { r, _, _ := send.Call(ctl[idList], 0x188, 0, 0); return int(int32(r)) }
func checked(id int) bool  { r, _, _ := send.Call(ctl[id], 0xF0, 0, 0); return r == 1 }
func check(id int, v bool) {
	var n uintptr
	if v {
		n = 1
	}
	send.Call(ctl[id], 0xF1, n, 0)
}
func getModInput() (string, string) {
	return strings.TrimSpace(textOf(idName)), filepath.Clean(strings.TrimSpace(textOf(idPath)))
}
func refreshList(sel int) {
	send.Call(ctl[idList], 0x184, 0, 0)
	var maxWidth int32
	dc, _, _ := user.NewProc("GetDC").Call(ctl[idList])
	if dc != 0 {
		defer user.NewProc("ReleaseDC").Call(ctl[idList], dc)
	}
	old := uintptr(0)
	if dc != 0 && font != 0 {
		old, _, _ = gdi.NewProc("SelectObject").Call(dc, font)
		if old != 0 {
			defer gdi.NewProc("SelectObject").Call(dc, old)
		}
	}
	for _, m := range config.Mods {
		value := wide(m.Name + "   —   " + m.Path)
		send.Call(ctl[idList], 0x180, 0, ptr(value))
		if dc != 0 {
			var size struct{ cx, cy int32 }
			encoded, _ := syscall.UTF16FromString(m.Name + "   —   " + m.Path)
			gdi.NewProc("GetTextExtentPoint32W").Call(dc, ptr(value), uintptr(len(encoded)-1), uintptr(unsafe.Pointer(&size)))
			if size.cx+12 > maxWidth {
				maxWidth = size.cx + 12
			}
		}
	}
	send.Call(ctl[idList], 0x194, uintptr(maxWidth), 0)
	if sel >= 0 && sel < len(config.Mods) {
		send.Call(ctl[idList], 0x186, uintptr(sel), 0)
		fillMod()
	} else {
		set(idName, "")
		set(idPath, "")
	}
}
func fillMod() {
	i := selected()
	if i >= 0 && i < len(config.Mods) {
		set(idName, config.Mods[i].Name)
		set(idPath, config.Mods[i].Path)
	}
}

func control(class, label string, id, x, y, w, h int, style uintptr) {
	v, _, _ := uCreate.Call(0, ptr(wide(class)), ptr(wide(label)), 0x50000000|style, uintptr(x), uintptr(y), uintptr(w), uintptr(h), hwnd, uintptr(id), instance, 0)
	lastControl = v
	if id != 0 {
		ctl[id] = v
	}
	if font != 0 {
		send.Call(v, 0x30, font, 1)
	}
}
func label(s string, x, y, w, h int) { control("STATIC", s, 0, x, y, w, h, 0) }
func styledLabel(s string, x, y, w, h int, f, color uintptr) {
	label(s, x, y, w, h)
	labelColors[lastControl] = color
	if f != 0 {
		send.Call(lastControl, 0x30, f, 1)
	}
}
func section(s string, x, y, w int)         { styledLabel(s, x, y, w, 25, sectionFont, 0x593722) }
func muted(s string, x, y, w, h int)        { styledLabel(s, x, y, w, h, font, 0x706254) }
func line(x, y, w int)                      { control("STATIC", "", 0, x, y, w, 2, 0x10) }
func edit(id, x, y, w, h int)               { control("EDIT", "", id, x, y, w, h, 0x00800000|0x10000|0x80) }
func button(s string, id, x, y, w, h int)   { control("BUTTON", s, id, x, y, w, h, 0x10000) }
func checkbox(s string, id, x, y, w, h int) { control("BUTTON", s, id, x, y, w, h, 0x10000|3) }
func buildControls() {
	styledLabel("HoI4 Mod Navigator", 24, 13, 590, 34, titleFont, 0x593722)
	muted("v2.0  /  Настройки интеграции", 24, 49, 740, 22)
	line(24, 78, 744)
	section("01   Оригинальная игра", 24, 87, 740)
	muted("Корневая папка, содержащая hoi4.exe", 24, 113, 740, 20)
	edit(idGame, 24, 136, 603, 30)
	button("Выбрать…", idBrowseGame, 640, 135, 128, 32)
	section("02   Мои моды", 24, 181, 740)
	muted("Выбери мод в списке для редактирования", 24, 207, 740, 20)
	control("LISTBOX", "", idList, 24, 231, 744, 77, 0x00800000|0x200000|0x100000|0x10000|1|0x100)

	label("Название", 24, 314, 220, 20)
	label("Корневая папка мода", 252, 314, 500, 20)
	edit(idName, 24, 337, 215, 30)
	edit(idPath, 252, 337, 375, 30)
	button("Выбрать…", idBrowseMod, 640, 336, 128, 32)
	button("Добавить мод", idAdd, 24, 376, 145, 32)
	button("Сохранить правки", idUpdate, 179, 376, 168, 32)
	button("Убрать из списка", idDelete, 357, 376, 168, 32)
	line(24, 420, 744)
	section("03   Команды Проводника", 24, 430, 740)
	checkbox("Открыть папку в моде", idOpen, 24, 460, 355, 25)
	checkbox("Скопировать файл в мод", idCopy, 398, 460, 370, 25)
	checkbox("Показать папку / файл в ванилле", idReverse, 24, 488, 355, 25)
	checkbox("Сравнить файлы (VS Code)", idCompare, 398, 488, 370, 25)
	checkbox("Создавать недостающие папки в моде", idCreate, 24, 516, 365, 25)
	checkbox("Фильтровать команды ПКМ", idFilter, 398, 516, 370, 25)
	checkbox("Совместимый режим (старое меню без DLL)", idLegacy, 24, 544, 744, 25)
	muted("По умолчанию: динамическое меню. Фильтр выше — только для старого режима.", 24, 574, 744, 22)
	line(24, 601, 744)
	control("BUTTON", "Применить настройки", idApply, 24, 613, 225, 36, 0x10000|1)
	button("Удалить интеграцию", idRemove, 260, 613, 205, 36)
	button("Закрыть", idClose, 640, 613, 128, 36)
	control("STATIC", "", idStatus, 24, 658, 744, 42, 0)
	labelColors[ctl[idStatus]] = 0x706254
	set(idGame, config.Game)
	refreshList(-1)
	check(idOpen, config.Open)
	check(idCopy, config.Copy)
	check(idReverse, config.Reverse)
	check(idCompare, config.Compare)
	check(idCreate, config.Create)
	check(idFilter, config.Filter)
	check(idLegacy, config.LegacyMenu)
	status("Укажи папки, добавь мод и нажми «Применить настройки».")
}
func collect() Config {
	c := config
	c.Game = filepath.Clean(strings.TrimSpace(textOf(idGame)))
	c.Open = checked(idOpen)
	c.Copy = checked(idCopy)
	c.Reverse = checked(idReverse)
	c.Compare = checked(idCompare)
	c.Create = checked(idCreate)
	c.Filter = checked(idFilter)
	c.LegacyMenu = checked(idLegacy)
	return c
}
func command(id int, notify int) {
	switch id {
	case idList:
		if notify == 1 {
			fillMod()
		}
	case idBrowseGame:
		if p, ok := browse("Выбери папку оригинальной Hearts of Iron IV", textOf(idGame)); ok {
			set(idGame, p)
		}
	case idBrowseMod:
		if p, ok := browse("Выбери корневую папку мода", textOf(idPath)); ok {
			set(idPath, p)
			if strings.TrimSpace(textOf(idName)) == "" {
				set(idName, filepath.Base(p))
			}
		}
	case idAdd, idUpdate:
		n, p := getModInput()
		if n == "" || p == "." {
			message(hwnd, "Укажи название и папку мода", 0x30)
			return
		}
		if e := rootCheck(p); e != nil {
			message(hwnd, e.Error(), 0x30)
			return
		}
		i := selected()
		if id == idUpdate && (i < 0 || i >= len(config.Mods)) {
			message(hwnd, "Сначала выбери мод в списке", 0x30)
			return
		}
		if id == idAdd && len(config.Mods) >= 20 {
			message(hwnd, "Максимум 20 модов", 0x30)
			return
		}
		for j, m := range config.Mods {
			if id == idUpdate && j == i {
				continue
			}
			if strings.EqualFold(m.Name, n) || inside(m.Path, p) || inside(p, m.Path) {
				message(hwnd, "Название или папка пересекается с другим модом", 0x30)
				return
			}
		}
		if id == idAdd {
			config.Mods = append(config.Mods, Mod{newID(), n, p})
			i = len(config.Mods) - 1
		} else {
			config.Mods[i].Name = n
			config.Mods[i].Path = p
		}
		refreshList(i)
		status("Изменения не применены. Нажми «Применить настройки».")
	case idDelete:
		i := selected()
		if i < 0 || i >= len(config.Mods) {
			return
		}
		if !ask("Убрать «" + config.Mods[i].Name + "» из списка?\nФайлы мода не будут удалены.") {
			return
		}
		config.Mods = append(config.Mods[:i], config.Mods[i+1:]...)
		refreshList(-1)
		status("Изменения не применены. Нажми «Применить настройки».")
	case idApply:
		c := collect()
		if e := validate(c); e != nil {
			message(hwnd, e.Error(), 0x30)
			return
		}
		if !c.Open && !c.Copy && !c.Reverse && !c.Compare {
			message(hwnd, "Включи хотя бы одну команду или удали интеграцию", 0x30)
			return
		}
		if e := installSelf(); e != nil {
			message(hwnd, "Не удалось сохранить EXE: "+e.Error(), 0x10)
			return
		}
		if e := saveConfig(c); e != nil {
			message(hwnd, "Не удалось сохранить настройки: "+e.Error(), 0x10)
			return
		}
		if e := registerMenus(c); e != nil {
			message(hwnd, "Настройки сохранены, но не удалось зарегистрировать меню: "+e.Error()+"\nПопробуй применить повторно.", 0x10)
			return
		}
		config = c
		status("Готово: команды ПКМ зарегистрированы.")
		if c.LegacyMenu {
			message(hwnd, "Настройки применены. Включён совместимый режим.", 0x40)
		} else {
			message(hwnd, "Настройки применены. Динамическое меню включено.\nЕсли меню не обновилось, перезапусти Проводник через Диспетчер задач.\nПри проблемах включи совместимый режим и примени настройки.", 0x40)
		}
	case idRemove:
		if !ask("Удалить команды из Проводника?\nНастройки и файлы игры/модов останутся.") {
			return
		}
		if e := removeMenus(); e != nil {
			message(hwnd, e.Error(), 0x10)
			return
		}
		status("Интеграция удалена. Настройки сохранены.")
		message(hwnd, "Пункты контекстного меню удалены. Для полного удаления после закрытия программы можно удалить:\n"+appDir(), 0x40)
	case idClose:
		user.NewProc("DestroyWindow").Call(hwnd)
	}
}

type wndClass struct {
	size, style                                uint32
	proc                                       uintptr
	clsExtra, wndExtra                         int32
	inst, icon, cursor, bg, menu, name, iconSm uintptr
}
type msg struct {
	hwnd           uintptr
	message        uint32
	wparam, lparam uintptr
	time           uint32
	x, y           int32
	private        uint32
}

var choosing bool
var choiceIndex = -1

func chooseDestination(choices []menuItem) int {
	choosing = true
	choiceIndex = -1
	showWindow("HoI4 — выбор мода", 460, 420, func() {
		styledLabel("Выбери мод", 16, 12, 430, 34, titleFont, 0x593722)
		muted("Где открыть соответствующую папку?", 16, 48, 430, 22)
		control("LISTBOX", "", idList, 16, 78, 430, 229, 0x00800000|0x200000|0x10000|1|0x100)
		for _, item := range choices {
			send.Call(ctl[idList], 0x180, 0, ptr(wide(item.label)))
		}
		send.Call(ctl[idList], 0x186, 0, 0)
		button("Открыть", idApply, 16, 320, 190, 32)
		button("Отмена", idClose, 256, 320, 190, 32)
	})
	choosing = false
	if choiceIndex >= len(choices) {
		return -1
	}
	return choiceIndex
}

func wndProc(h uintptr, m uint32, w, l uintptr) uintptr {
	switch m {
	case 0x138: // WM_CTLCOLORSTATIC: consistent text, respecting system background.
		if color, ok := labelColors[l]; ok && !highContrast {
			gdi.NewProc("SetTextColor").Call(w, color)
		}
		gdi.NewProc("SetBkColor").Call(w, panelColor)
		return panelBrush
	case 0x111:
		id, notify := int(w&0xffff), int((w>>16)&0xffff)
		if choosing {
			if id == idApply || (id == idList && notify == 2) {
				if i := selected(); i >= 0 {
					choiceIndex = i
					user.NewProc("DestroyWindow").Call(h)
				}
			} else if id == idClose || id == 2 {
				user.NewProc("DestroyWindow").Call(h)
			}
			return 0
		}
		command(id, notify)
		return 0
	case 0x10:
		user.NewProc("DestroyWindow").Call(h)
		return 0
	case 2:
		user.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	r, _, _ := def.Call(h, uintptr(m), w, l)
	return r
}
func runGUI() {
	if os.Getenv("LOCALAPPDATA") == "" {
		message(0, "Переменная LOCALAPPDATA не найдена", 0x10)
		return
	}
	var e error
	config, e = loadConfig()
	if e != nil {
		message(0, "Не удалось прочитать настройки:\n"+e.Error()+"\nПроверь файл: "+configPath(), 0x10)
		return
	}
	showWindow(appName+"  •  v2.0", 810, 750, buildControls)
}
func showWindow(title string, width, height int, build func()) {
	// DPI-unaware UI is automatically scaled by Windows, keeping the fixed layout usable.
	ole.NewProc("CoInitializeEx").Call(0, 2|4)
	defer ole.NewProc("CoUninitialize").Call()
	instance, _, _ = kernel.NewProc("GetModuleHandleW").Call(0)
	// Visual styles are enabled by the embedded common-controls v6 manifest.
	syscall.NewLazyDLL("comctl32.dll").NewProc("InitCommonControls").Call()
	createFont := func(height int32, weight uintptr) uintptr {
		f, _, _ := gdi.NewProc("CreateFontW").Call(uintptr(height), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 5, 0, ptr(wide("Segoe UI")))
		return f
	}
	font = createFont(-15, 400)
	sectionFont = createFont(-17, 600)
	titleFont = createFont(-28, 600)
	if font == 0 {
		font, _, _ = gdi.NewProc("GetStockObject").Call(17)
	} else {
		defer gdi.NewProc("DeleteObject").Call(font)
	}
	if sectionFont != 0 {
		defer gdi.NewProc("DeleteObject").Call(sectionFont)
	}
	if titleFont != 0 {
		defer gdi.NewProc("DeleteObject").Call(titleFont)
	}
	var hc struct {
		size, flags uint32
		scheme      uintptr
	}
	hc.size = uint32(unsafe.Sizeof(hc))
	user.NewProc("SystemParametersInfoW").Call(0x42, uintptr(hc.size), uintptr(unsafe.Pointer(&hc)), 0)
	highContrast = hc.flags&1 != 0
	if highContrast {
		panelBrush, _, _ = user.NewProc("GetSysColorBrush").Call(5)
		panelColor, _, _ = user.NewProc("GetSysColor").Call(5)
	} else {
		panelBrush, _, _ = gdi.NewProc("GetStockObject").Call(0)
	}

	cursor, _, _ := user.NewProc("LoadCursorW").Call(0, 32512)
	icon, _, _ := user.NewProc("LoadIconW").Call(instance, 1)
	class := wide("HoI4ModNavigatorWindow")
	wc := wndClass{size: uint32(unsafe.Sizeof(wndClass{})), proc: syscall.NewCallback(wndProc), inst: instance, icon: icon, cursor: cursor, bg: panelBrush, name: ptr(class), iconSm: icon}
	r, _, er := user.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		message(0, fmt.Sprint(er), 0x10)
		return
	}
	style := uintptr(0x00CA0000)
	hwnd, _, er = uCreate.Call(0, ptr(class), ptr(wide(title)), style, 0x80000000, 0x80000000, uintptr(width), uintptr(height), 0, 0, instance, 0)
	if hwnd == 0 {
		message(0, fmt.Sprint(er), 0x10)
		return
	}
	build()
	user.NewProc("ShowWindow").Call(hwnd, 5)
	user.NewProc("UpdateWindow").Call(hwnd)
	var m msg
	for {
		r, _, _ := user.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		handled, _, _ := user.NewProc("IsDialogMessageW").Call(hwnd, uintptr(unsafe.Pointer(&m)))
		if handled == 0 {
			user.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&m)))
			user.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&m)))
		}
	}
	hwnd = 0
}
