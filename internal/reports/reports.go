package reports

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
)

func Generate(params map[string]string, out io.Writer) error {
	name, ok := params["name"]
	if !ok || strings.TrimSpace(name) == "" {
		return fmt.Errorf("rep requiere -name")
	}
	outputPath, ok := params["path"]
	if !ok || strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("rep requiere -path")
	}
	id, ok := params["id"]
	if !ok || strings.TrimSpace(id) == "" {
		return fmt.Errorf("rep requiere -id")
	}

	reportName := strings.ToLower(name)
	if reportName == "bm_bloc" {
		reportName = "bm_block"
	}

	switch reportName {
	case "mbr", "disk":
		return generateDiskReport(reportName, outputPath, id, out)
	case "sb", "inode", "block", "tree":
		return generateExt2DotReport(reportName, outputPath, id, out)
	case "bm_inode", "bm_block":
		return generateBitmapReport(reportName, outputPath, id, out)
	case "file":
		target, ok := params["path_file_ls"]
		if !ok || strings.TrimSpace(target) == "" {
			return fmt.Errorf("rep file requiere -path_file_ls")
		}
		return generateFileReport(outputPath, id, target, out)
	case "ls":
		target, ok := params["path_file_ls"]
		if !ok || strings.TrimSpace(target) == "" {
			return fmt.Errorf("rep ls requiere -path_file_ls")
		}
		return generateLSDotReport(outputPath, id, target, out)
	default:
		return fmt.Errorf("reporte desconocido %q", reportName)
	}
}

func generateDiskReport(reportName string, outputPath string, id string, out io.Writer) error {
	mounted, ok := mount.Global.GetMounted(id)
	if !ok {
		return fmt.Errorf("no existe montaje con id %q", id)
	}
	mbr, err := disk.ReadMBR(mounted.DiskPath)
	if err != nil {
		return err
	}

	var dot string
	switch reportName {
	case "mbr":
		dot, err = BuildMBRDot(mounted.DiskPath, mbr)
	case "disk":
		dot, err = BuildDiskDot(mounted.DiskPath, mbr)
	}
	if err != nil {
		return err
	}

	dotPath, err := writeDotAndRender(outputPath, dot)
	if err != nil {
		fmt.Fprintf(out, "Advertencia: %v\n", err)
		return nil
	}

	fmt.Fprintf(out, "Reporte %s generado. DOT: %s\n", reportName, dotPath)
	return nil
}

func generateExt2DotReport(reportName string, outputPath string, id string, out io.Writer) error {
	ctx, file, err := LoadExt2Context(id)
	if err != nil {
		return err
	}
	defer file.Close()

	var dot string
	switch reportName {
	case "sb":
		dot = BuildSuperBlockDot(ctx.SuperBlock)
	case "inode":
		dot, err = BuildInodeDot(file, ctx.SuperBlock)
	case "block":
		dot, err = BuildBlockDot(file, ctx.SuperBlock)
	case "tree":
		dot, err = BuildTreeDot(file, ctx.SuperBlock)
	}
	if err != nil {
		return err
	}
	dotPath, err := writeDotAndRender(outputPath, dot)
	if err != nil {
		fmt.Fprintf(out, "Advertencia: %v\n", err)
		return nil
	}
	fmt.Fprintf(out, "Reporte %s generado. DOT: %s\n", reportName, dotPath)
	return nil
}

func generateBitmapReport(reportName string, outputPath string, id string, out io.Writer) error {
	ctx, file, err := LoadExt2Context(id)
	if err != nil {
		return err
	}
	defer file.Close()
	var bitmap []byte
	if reportName == "bm_inode" {
		bitmap, err = ReadInodeBitmapForReport(file, ctx.SuperBlock)
	} else {
		bitmap, err = ReadBlockBitmapForReport(file, ctx.SuperBlock)
	}
	if err != nil {
		return err
	}
	if strings.EqualFold(filepath.Ext(outputPath), ".txt") {
		if err := writePlainText(outputPath, BuildBitmapText(bitmap)); err != nil {
			return err
		}
		fmt.Fprintf(out, "Reporte %s generado: %s\n", reportName, outputPath)
		return nil
	}
	dotPath, err := writeDotAndRender(outputPath, BuildBitmapDot(reportName, bitmap))
	if err != nil {
		fmt.Fprintf(out, "Advertencia: %v\n", err)
		return nil
	}
	fmt.Fprintf(out, "Reporte %s generado. DOT: %s\n", reportName, dotPath)
	return nil
}

func generateFileReport(outputPath string, id string, target string, out io.Writer) error {
	ctx, file, err := LoadExt2Context(id)
	if err != nil {
		return err
	}
	defer file.Close()
	content, err := BuildFileReport(ctx, file, target)
	if err != nil {
		return err
	}
	if err := writePlainText(outputPath, content); err != nil {
		return err
	}
	fmt.Fprintf(out, "Reporte file generado: %s\n", outputPath)
	return nil
}

func generateLSDotReport(outputPath string, id string, target string, out io.Writer) error {
	ctx, file, err := LoadExt2Context(id)
	if err != nil {
		return err
	}
	defer file.Close()
	dot, err := BuildLSDot(file, ctx.SuperBlock, target)
	if err != nil {
		return err
	}
	dotPath, err := writeDotAndRender(outputPath, dot)
	if err != nil {
		fmt.Fprintf(out, "Advertencia: %v\n", err)
		return nil
	}
	fmt.Fprintf(out, "Reporte ls generado. DOT: %s\n", dotPath)
	return nil
}
