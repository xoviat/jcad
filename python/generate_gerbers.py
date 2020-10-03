"""
The purpose of this script is to generate gerbers, given a pcb file


usage: generate_gerbers.py <pcb_file.pcb> <target_folder>
"""

import sys

from pcbnew import LoadBoard, GERBER_WRITER

filename = sys.argv[1]

pcb = LoadBoard(filename)
gerber_writer = GERBER_WRITER(pcb)

gerber_writer.SetFormat(self, aRightDigits=6)
gerber_writer.CreateDrillandMapFilesSet(
    aPlotDirectory,
    aGenDrill,
    aGenMap,
    aReporter=None,
)


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
        help="output in .brd format",
    )

    args = parser.parse_args()
    convert(pcbnew.LoadBoard(args.kicad_pcb_file), args.brd_file)


if __name__ == "__main__":
    main()