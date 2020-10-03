"""
The purpose of this script is to generate a cpl file, given a pcb file


usage: generate_cpl.py <pcb_file.pcb> <cpl_file.cpl>
"""

import sys
import re
import argparse

import pcbnew


def skip_module(module, tp=False):
    refdes = module.GetReference()
    if refdes == "REF**":
        return True
    if tp and not refdes.startswith("TP"):
        return True
    if not tp and refdes.startswith("TP"):
        return True
    return False


def coord(nanometers):
    milliinches = nanometers * 5 // 127000
    return milliinches


def y_coord(obj, maxy, y):
    if obj.IsFlipped():
        return coord(maxy - y)
    else:
        return coord(y)


def pad_sort_key(name):
    if re.match(r"^\d+$", name):
        return (0, int(name))
    else:
        return (1, name)


def convert(pcb, brd):
    IU_PER_MM = PCB_IU_PER_MM = 1e6
    IU_PER_MILS = (IU_PER_MM * 0.0254)

    conv_unit_inch = 0.001 / IU_PER_MILS
    conv_unit_mm = 1.0 / IU_PER_MM

    units = pcbnew.GetUserUnits()

    # TODO: select units from board somehow
    conv_unit = conv_unit_mm

    # Board outline
    outlines = pcbnew.SHAPE_POLY_SET()
    pcb.GetBoardPolygonOutlines(outlines, "")
    outline = outlines.Outline(0)
    outline_points = [outline.Point(n) for n in range(outline.PointCount())]
    outline_maxx = max(map(lambda p: p.x, outline_points))
    outline_maxy = max(map(lambda p: p.y, outline_points))
    m_place_offset = pcb.GetDesignSettings().m_AuxOrigin

    # Parts
    module_list = pcb.GetModules()
    modules = []
    while module_list:
        if not skip_module(module_list):
            modules.append(module_list)
        module_list = module_list.Next()

    pin_at = 0

    # Logic taken from pcbnew/exporters/export_footprints_placefile.cpp
    # See https://gitlab.com/kicad/code/kicad/-/issues/2453
    for module in modules:
        footprint_pos = module.GetPosition()
        footprint_pos -= m_place_offset

        layer = module.GetLayer()
        if(layer not in (pcbnew.B_Cu, pcbnew.F_Cu)):
            continue

        if(layer == pcbnew.B_Cu):
            footprint_pos.x = - footprint_pos.x

        module_bbox = module.GetBoundingBox()
        brd.write(
            "{ref}	{value}	{name}	{x0}	{y0}	{orientation}	{layer}\n".format(
                ref=module.GetReference(),
                value=module.GetValue(),
                name=module.GetFPID().GetLibItemName(),
                x0=footprint_pos.x * conv_unit,
                y0=-footprint_pos.y * conv_unit,
                orientation=module.GetOrientation() / 10.0,
                layer="top" if layer == pcbnew.F_Cu else "bottom",
            )
        )
        pin_at += module.GetPadCount()
    brd.write("\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "kicad_pcb_file",
        metavar="KICAD-PCB-FILE",
        type=str,
        help="input in .kicad_pcb format",
    )
    parser.add_argument(
        "brd_file",
        metavar="BRD-FILE",
        type=argparse.FileType("wt"),
        help="output in .cpl format",
    )

    args = parser.parse_args()
    convert(pcbnew.LoadBoard(args.kicad_pcb_file), args.brd_file)


if __name__ == "__main__":
    main()
