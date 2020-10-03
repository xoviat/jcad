# JCAD

This project aims to simplify and clarify the process of ordering SMT assembly boards from JLCPCB.
The problem is as such: given a KiCAD .sch and .kicad_pcb file, what is the simpliest and fastest way
to order SMT assembled boards from JLCPCB? The current process is as follows:

1. Generate a BOM in EEschema
2. Generate a POS file in PCBnew
3. Generate Gerbers in PCBnew
4. Run [tooling](https://github.com/wokwi/kicad-jlcpcb-bom-plugin) to transform the BOM and pos file
5. Remove thorugh-hole components from the BOM
6. Upload the files to JLCPCB and order the SMT board
7. Order additional through-hole components from a distributor, and solder these components yourself

## What can't change

1. Though-hole components and components not available from JLCPCB must be be ordered and soldered yourself.
2. As of now, we can't access the JLCPCB API, so we still have to click through their wizard

## What can change

1. The input files to the JLCPCB website can be directly derived from the KiCAD .sch and .kicad_pcb files
2. The components can be assigned interactively, before uploading to JLCPCB.
3. Calculations can be done offline, using the JLCPCB component database.
4. Footprints can be assigned offline, using the JLCPCB component database, directly editing the .sch file.

## What our envisionsed workflow looks like

1. After creating a schema, open our tool and assign footprints directly using  the JLCPCB component database, and directly assign the JLCPCB part number.
2. Then, after creating the PCB, open our tool again, and directly generate the JLCPCB files and validate component orientation offline.
3. In addition, directly generate an additional BOM of through-hole parts not included in the SMT assembly that the user can upload to Mouser. Validate using the Mouser API.

## Why this is possible

1. The .sch format is not too complex and we only need to edit properties of the components. We will not edit component placement or wires.
2. The .kicad_pcb file can be read using the pcbnew Python API. We can generate the required outputs and postprocess them ourselves.
3. The BOM and POS files are simple to read and should not pose too many difficulties.

## Libraries

1. I golang because it's performant, compiles quickly, and executes natively.
2. I use lxn/walk because I avoid cgo at all costs. Unfortunatley, this restricts the gui of this application to windows, but command-line could
   be developed for other plafforms, and I am not opposed to other UI frameworks if they avoid cgo. I will focus on windows because it's what
   I use.
3. This applicadtion will require KiCad to be installed to access the pcbnew api. It will use python scripts to access the pcbnew API.

## Issues

Feel free to give ideas in the issues. This project solves a real problem, and therefore it might (hopefully) become popular.
