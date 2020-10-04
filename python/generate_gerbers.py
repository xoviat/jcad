"""
The purpose of this script is to generate gerbers, given a pcb file


usage: generate_gerbers.py <pcb_file.pcb> <target_folder>
"""

import os
import sys
import re
import argparse
import pcbnew


def convert(pcb, dir):
    board = pcb
    plotDir = os.path.abspath(dir)

    # prepare the gerber job file
    gen_job_file = False

    pctl = pcbnew.PLOT_CONTROLLER(board)

    popt = pctl.GetPlotOptions()

    popt.SetOutputDirectory(plotDir)

    # Set some important plot options (see pcb_plot_params.h):
    popt.SetPlotFrameRef(False)  # do not change it
    popt.SetLineWidth(pcbnew.FromMM(0.35))

    popt.SetAutoScale(False)  # do not change it
    popt.SetScale(1)  # do not change it
    popt.SetMirror(False)
    popt.SetUseGerberAttributes(True)
    popt.SetIncludeGerberNetlistInfo(True)
    popt.SetCreateGerberJobFile(gen_job_file)
    popt.SetUseGerberProtelExtensions(False)
    popt.SetExcludeEdgeLayer(False)
    popt.SetScale(1)
    popt.SetUseAuxOrigin(True)

    # This by gerbers only
    popt.SetSubtractMaskFromSilk(False)
    # Disable plot pad holes
    popt.SetDrillMarksType(pcbnew.PCB_PLOT_PARAMS.NO_DRILL_SHAPE)
    # Skip plot pad NPTH when possible: when drill size and shape == pad size and shape
    # usually sel to True for copper layers
    popt.SetSkipPlotNPTH_Pads(False)

    # prepare the gerber job file
    jobfile_writer = pcbnew.GERBER_JOBFILE_WRITER(board)

    # Once the defaults are set it become pretty easy...
    # I have a Turing-complete programming language here: I'll use it...
    # param 0 is a string added to the file base name to identify the drawing
    # param 1 is the layer ID
    # param 2 is a comment
    plot_plan = [
        ("F_Cu", pcbnew.F_Cu, "Top layer"),
        ("B_Cu", pcbnew.B_Cu, "Bottom layer"),
        ("B_Paste", pcbnew.B_Paste, "Paste Bottom"),
        ("F_Paste", pcbnew.F_Paste, "Paste top"),
        ("F_SilkS", pcbnew.F_SilkS, "Silk top"),
        ("B_SilkS", pcbnew.B_SilkS, "Silk top"),
        ("B_Mask", pcbnew.B_Mask, "Mask bottom"),
        ("F_Mask", pcbnew.F_Mask, "Mask top"),
        ("Edge_Cuts", pcbnew.Edge_Cuts, "Edges"),
    ]

    for layer_info in plot_plan:
        if layer_info[1] <= pcbnew.B_Cu:
            popt.SetSkipPlotNPTH_Pads(True)
        else:
            popt.SetSkipPlotNPTH_Pads(False)

        pctl.SetLayer(layer_info[1])
        pctl.OpenPlotfile(
            layer_info[0], pcbnew.PLOT_FORMAT_GERBER, layer_info[2])
        print 'plot %s' % pctl.GetPlotFileName()
        if gen_job_file == True:
            jobfile_writer.AddGbrFile(
                layer_info[1], os.path.basename(pctl.GetPlotFileName()))
        if pctl.PlotLayer() == False:
            print "plot error"

    # generate internal copper layers, if any
    lyrcnt = board.GetCopperLayerCount()

    for innerlyr in range(1, lyrcnt-1):
        popt.SetSkipPlotNPTH_Pads(True)
        pctl.SetLayer(innerlyr)
        lyrname = 'In%s_Cu' % innerlyr
        pctl.OpenPlotfile(lyrname, pcbnew.PLOT_FORMAT_GERBER, "inner")
        print 'plot %s' % pctl.GetPlotFileName()
        if pctl.PlotLayer() == False:
            print "plot error"

    # At the end you have to close the last plot, otherwise you don't know when
    # the object will be recycled!
    pctl.ClosePlot()

    # Fabricators need drill files.
    # sometimes a drill map file is asked (for verification purpose)
    drlwriter = pcbnew.EXCELLON_WRITER(board)
    drlwriter.SetMapFileFormat(pcbnew.PLOT_FORMAT_PDF)

    mirror = False
    minimalHeader = False
    offset = pcbnew.wxPoint(0, 0)
    # False to generate 2 separate drill files (one for plated holes, one for non plated holes)
    # True to generate only one drill file
    mergeNPTH = True
    drlwriter.SetOptions(mirror, minimalHeader, offset, mergeNPTH)

    metricFmt = True
    drlwriter.SetFormat(metricFmt)

    genDrl = True
    genMap = True
    print 'create drill and map files in %s' % pctl.GetPlotDirName()
    drlwriter.CreateDrillandMapFilesSet(pctl.GetPlotDirName(), genDrl, genMap)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "kicad_pcb_file",
        metavar="KICAD-PCB-FILE",
        type=str,
        help="input in .kicad_pcb format",
    )
    parser.add_argument(
        "gerber_dir",
        metavar="GERBER-DIR",
        type=str,
        help="output in gerber format",
    )

    args = parser.parse_args()
    convert(pcbnew.LoadBoard(args.kicad_pcb_file), args.gerber_dir)


if __name__ == "__main__":
    main()
