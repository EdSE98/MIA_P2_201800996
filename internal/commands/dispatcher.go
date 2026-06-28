package commands

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/reports"
	"mia_p1_201800996/internal/session"
	"mia_p1_201800996/internal/users"
)

type Dispatcher struct {
	in         io.Reader
	out        io.Writer
	specs      map[string]commandSpec
	scriptMode bool
}

type commandSpec struct {
	params     map[string]bool
	flags      map[string]bool
	dynamicCat bool
}

func NewDispatcher(in io.Reader, out io.Writer) *Dispatcher {
	return &Dispatcher{
		in:    in,
		out:   out,
		specs: buildSpecs(),
	}
}

func (d *Dispatcher) SetScriptMode(enabled bool) {
	d.scriptMode = enabled
}

func (d *Dispatcher) Execute(cmd Command) (bool, error) {
	if cmd.IsComment() {
		fmt.Fprintln(d.out, cmd.Raw)
		return false, nil
	}

	spec, ok := d.specs[cmd.Name]
	if !ok {
		return false, fmt.Errorf("linea %d: comando desconocido %q", cmd.Line, cmd.Name)
	}

	for key := range cmd.Params {
		if !spec.allowsParam(key) {
			return false, fmt.Errorf("linea %d: parametro desconocido para %s: -%s", cmd.Line, cmd.Name, key)
		}
	}

	for key := range cmd.Flags {
		if !spec.flags[key] {
			return false, fmt.Errorf("linea %d: flag desconocido para %s: -%s", cmd.Line, cmd.Name, key)
		}
	}

	switch cmd.Name {
	case "mkdisk":
		if err := disk.MakeDiskFromParams(cmd.Params); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Disco creado: %s\n", cmd.Params["path"])
		return false, nil
	case "rmdisk":
		if err := d.removeDisk(cmd.Params); err != nil {
			return false, err
		}
		return false, nil
	case "fdisk":
		_, hasAdd := cmd.Params["add"]
		_, hasDelete := cmd.Params["delete"]
		if hasAdd && hasDelete {
			return false, fmt.Errorf("fdisk no permite usar -add y -delete al mismo tiempo")
		}
		if (hasAdd || hasDelete) && mount.Global.IsMounted(cmd.Params["path"], cmd.Params["name"]) {
			return false, fmt.Errorf("no se puede modificar una particion montada")
		}
		switch {
		case hasAdd:
			if err := partition.ResizeFromParams(cmd.Params); err != nil {
				return false, err
			}
			fmt.Fprintf(d.out, "Particion redimensionada: %s\n", cmd.Params["name"])
		case hasDelete:
			if err := partition.DeleteFromParams(cmd.Params); err != nil {
				return false, err
			}
			fmt.Fprintf(d.out, "Particion eliminada: %s\n", cmd.Params["name"])
		default:
			if err := partition.CreateFromParams(cmd.Params); err != nil {
				return false, err
			}
			fmt.Fprintf(d.out, "Particion creada: %s\n", cmd.Params["name"])
		}
		return false, nil
	case "mount":
		mounted, err := mount.Global.Mount(cmd.Params["path"], cmd.Params["name"])
		if err != nil {
			return false, err
		}
		if mounted.PartitionType == 'E' {
			fmt.Fprintf(d.out, "Advertencia: montando particion extendida %s\n", mounted.PartitionName)
		}
		fmt.Fprintf(d.out, "Particion montada con ID: %s\n", mounted.ID)
		d.printMountedPartitions()
		return false, nil
	case "unmount":
		id := cmd.Params["id"]
		if err := mount.Global.Unmount(id); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Particion desmontada: %s\n", id)
		if session.ClearIfMountedID(id) {
			fmt.Fprintln(d.out, "Advertencia: se cerro la sesion activa porque se desmontó la particion")
		}
		return false, nil
	case "mkfs":
		if err := fs.FormatFromParams(cmd.Params, d.out); err != nil {
			return false, err
		}
		return false, nil
	case "login":
		logged, err := session.Login(cmd.Params["user"], cmd.Params["pass"], cmd.Params["id"])
		if err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Sesion iniciada: %s en %s\n", logged.User, logged.MountedID)
		return false, nil
	case "logout":
		if err := session.Logout(); err != nil {
			return false, err
		}
		fmt.Fprintln(d.out, "Sesion cerrada")
		return false, nil
	case "mkgrp":
		if err := d.updateUsersFile(func(content string) (string, string, error) {
			next, err := users.MakeGroup(content, cmd.Params["name"])
			return next, fmt.Sprintf("Grupo creado: %s", cmd.Params["name"]), err
		}); err != nil {
			return false, err
		}
		return false, nil
	case "rmgrp":
		if err := d.updateUsersFile(func(content string) (string, string, error) {
			next, err := users.RemoveGroup(content, cmd.Params["name"])
			return next, fmt.Sprintf("Grupo eliminado: %s", cmd.Params["name"]), err
		}); err != nil {
			return false, err
		}
		return false, nil
	case "mkusr":
		if err := d.updateUsersFile(func(content string) (string, string, error) {
			next, err := users.MakeUser(content, cmd.Params["user"], cmd.Params["pass"], cmd.Params["grp"])
			return next, fmt.Sprintf("Usuario creado: %s", cmd.Params["user"]), err
		}); err != nil {
			return false, err
		}
		return false, nil
	case "rmusr":
		if err := d.updateUsersFile(func(content string) (string, string, error) {
			next, err := users.RemoveUser(content, cmd.Params["user"])
			return next, fmt.Sprintf("Usuario eliminado: %s", cmd.Params["user"]), err
		}); err != nil {
			return false, err
		}
		return false, nil
	case "chgrp":
		if err := d.updateUsersFile(func(content string) (string, string, error) {
			next, err := users.ChangeUserGroup(content, cmd.Params["user"], cmd.Params["grp"])
			return next, fmt.Sprintf("Grupo de usuario actualizado: %s -> %s", cmd.Params["user"], cmd.Params["grp"]), err
		}); err != nil {
			return false, err
		}
		return false, nil
	case "mkdir":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		if err := fs.Mkdir(active.DiskPath, int64(active.PartitionStart), cmd.Params["path"], cmd.Flags["p"], actor); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Carpeta creada: %s\n", cmd.Params["path"])
		return false, nil
	case "mkfile":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		size := int64(0)
		if value, ok := cmd.Params["size"]; ok && value != "" {
			size, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return false, fmt.Errorf("size invalido %q", value)
			}
		}
		if err := fs.Mkfile(active.DiskPath, int64(active.PartitionStart), cmd.Params["path"], cmd.Flags["r"], size, cmd.Params["cont"], actor); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Archivo creado: %s\n", cmd.Params["path"])
		return false, nil
	case "cat":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		content, err := fs.Cat(active.DiskPath, int64(active.PartitionStart), cmd.Params, actor)
		if err != nil {
			return false, err
		}
		fmt.Fprintln(d.out, content)
		return false, nil
	case "edit":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		if err := fs.Edit(active.DiskPath, int64(active.PartitionStart), cmd.Params["path"], cmd.Params["contenido"], actor); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Archivo editado: %s\n", cmd.Params["path"])
		return false, nil
	case "rename":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		if err := fs.Rename(active.DiskPath, int64(active.PartitionStart), cmd.Params["path"], cmd.Params["name"], actor); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Archivo o carpeta renombrado: %s\n", cmd.Params["path"])
		return false, nil
	case "remove":
		active, actor, err := activeFSActor()
		if err != nil {
			return false, err
		}
		if err := fs.Remove(active.DiskPath, int64(active.PartitionStart), cmd.Params["path"], actor); err != nil {
			return false, err
		}
		fmt.Fprintf(d.out, "Archivo o carpeta eliminado: %s\n", cmd.Params["path"])
		return false, nil
	case "rep":
		if err := reports.Generate(cmd.Params, d.out); err != nil {
			return false, err
		}
		return false, nil
	case "pause":
		return false, d.pause()
	case "exit":
		fmt.Fprintln(d.out, "Saliendo...")
		return true, nil
	default:
		if requiresSession(cmd.Name) {
			if _, err := session.RequireActive(); err != nil {
				return false, err
			}
		}
		fmt.Fprintf(d.out, "[STUB] %s recibido con params: %s flags: %s\n", cmd.Name, formatParams(cmd.Params), formatFlags(cmd.Flags))
		return false, nil
	}
}

func (d *Dispatcher) printMountedPartitions() {
	mounted := mount.Global.List()
	fmt.Fprintln(d.out, "Particiones montadas:")
	fmt.Fprintln(d.out, "ID\tDisco\tParticion")
	for _, item := range mounted {
		fmt.Fprintf(d.out, "%s\t%s\t%s\n", item.ID, item.DiskPath, item.PartitionName)
	}
}

func activeFSActor() (session.Session, fs.Actor, error) {
	active, err := session.RequireActive()
	if err != nil {
		return session.Session{}, fs.Actor{}, err
	}
	return active, fs.Actor{User: active.User, UID: active.UID, GID: active.GID}, nil
}

func (d *Dispatcher) updateUsersFile(operation func(content string) (string, string, error)) error {
	active, err := session.RequireRoot()
	if err != nil {
		return err
	}
	file, _, err := disk.OpenReadWrite(active.DiskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	sb, err := fs.ReadSuperBlock(file, int64(active.PartitionStart))
	if err != nil {
		return err
	}
	if sb.SMagic != 0xEF53 || sb.SFilesystemType != 2 {
		return fmt.Errorf("la particion no esta formateada como EXT2")
	}
	content, err := fs.ReadRootUsersFile(file, sb)
	if err != nil {
		return err
	}
	nextContent, message, err := operation(content)
	if err != nil {
		return err
	}
	if _, err := fs.WriteRootUsersFile(file, sb, nextContent); err != nil {
		return err
	}
	fmt.Fprintln(d.out, message)
	return nil
}

func (d *Dispatcher) removeDisk(params map[string]string) error {
	path, ok := params["path"]
	if !ok {
		return fmt.Errorf("rmdisk requiere -path")
	}

	if d.scriptMode {
		fmt.Fprintf(d.out, "Eliminando disco en modo script: %s\n", path)
	} else {
		confirmed, err := d.confirmRemove()
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(d.out, "Eliminacion cancelada")
			return nil
		}
	}

	if err := disk.RemoveDisk(path); err != nil {
		return err
	}
	mount.Global.UnmountByDisk(path)
	if session.ClearIfDiskPath(path) {
		fmt.Fprintln(d.out, "Advertencia: se cerro la sesion activa porque se eliminó el disco")
	}
	fmt.Fprintf(d.out, "Disco eliminado: %s\n", path)
	return nil
}

func (d *Dispatcher) confirmRemove() (bool, error) {
	fmt.Fprint(d.out, "¿Está seguro que desea eliminar el disco? (s/n): ")
	reader := bufio.NewReader(d.in)
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "s" || answer == "si" || answer == "y" || answer == "yes", nil
}

func (d *Dispatcher) pause() error {
	fmt.Fprint(d.out, "Presione ENTER para continuar...")
	reader := bufio.NewReader(d.in)
	_, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	fmt.Fprintln(d.out)
	return nil
}

func (s commandSpec) allowsParam(key string) bool {
	if s.params[key] {
		return true
	}
	if !s.dynamicCat {
		return false
	}
	if key == "file" {
		return true
	}
	if !strings.HasPrefix(key, "file") {
		return false
	}
	suffix := strings.TrimPrefix(key, "file")
	if suffix == "" {
		return true
	}
	for _, ch := range suffix {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func buildSpecs() map[string]commandSpec {
	return map[string]commandSpec{
		"mkdisk":  spec([]string{"size", "fit", "unit", "path"}, nil),
		"rmdisk":  spec([]string{"path"}, nil),
		"fdisk":   spec([]string{"size", "unit", "path", "type", "fit", "name", "add", "delete"}, nil),
		"mount":   spec([]string{"path", "name"}, nil),
		"unmount": spec([]string{"id"}, nil),
		"mkfs":    spec([]string{"id", "type"}, nil),
		"login":   spec([]string{"user", "pass", "id"}, nil),
		"logout":  spec(nil, nil),
		"mkgrp":   spec([]string{"name"}, nil),
		"rmgrp":   spec([]string{"name"}, nil),
		"mkusr":   spec([]string{"user", "pass", "grp"}, nil),
		"rmusr":   spec([]string{"user"}, nil),
		"chgrp":   spec([]string{"user", "grp"}, nil),
		"mkdir":   spec([]string{"path"}, []string{"p"}),
		"mkfile":  spec([]string{"path", "size", "cont"}, []string{"r"}),
		"edit":    spec([]string{"path", "contenido"}, nil),
		"rename":  spec([]string{"path", "name"}, nil),
		"remove":  spec([]string{"path"}, nil),
		"cat": {
			params:     map[string]bool{"file": true},
			flags:      map[string]bool{},
			dynamicCat: true,
		},
		"rep":   spec([]string{"name", "path", "id", "path_file_ls"}, nil),
		"pause": spec(nil, nil),
		"exit":  spec(nil, nil),
	}
}

func requiresSession(command string) bool {
	switch command {
	case "mkgrp", "rmgrp", "mkusr", "rmusr", "chgrp", "mkdir", "mkfile", "cat", "edit", "rename", "remove":
		return true
	default:
		return false
	}
}

func spec(params []string, flags []string) commandSpec {
	return commandSpec{
		params: stringSet(params),
		flags:  stringSet(flags),
	}
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func formatParams(params map[string]string) string {
	if len(params) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", key, params[key]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func formatFlags(flags map[string]bool) string {
	if len(flags) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(flags))
	for key := range flags {
		if flags[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return "{" + strings.Join(keys, ", ") + "}"
}
