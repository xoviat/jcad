# JCAD: A JLCPCB Manufacturing File Generator for KiCad

JCAD is a program to generate manufacturing files for JLCPCB directly from a `.kicad_pcb` file.
In contrast to other approaches, JCAD does not integrate with KiCad directly or require modification
of schematic or PCB design files in order to generate outputs. Instead, JCAD relies on a consistent
component naming scheme that allows the program to generate manufacturing files from a PCB file in
seconds once the configuration has been established.

## How JCAD works

JCAD associates each combination of designator, footprint, and value on a KiCad PCB with a specific
JLCPCB part. JCAD uses KiCad's new builtin command line client to extract data from the KiCad files
then matches this data with its internal database to prepare manufacturing files.

To reduce designer configuration, JCAD relies on a strict naming scheme for basic parts, and a more
flexible naming scheme for extended parts:

- Basic capacitors, inductors, and resistors must be named unitless
    * Example: 10u, 10k, 10p
- Numerical component values must be equal to or greater than 1, and must use decimal format
    * Example: 10.1k, 10.5p
- Designators must be labeled as follows:
    * Inductors, ferrite beads, and common-mode chokes must be labeled with the L designator
    * Transformers must be labeled with the T designator
    * Crystals must be labeled with the Y designator
- Extended components, including capacitors, inductors, and resistors, and all other basic components must be given the manufacturer name

## Generating Manufacturing Files

The core of JCAD is the `jcad generate` command. The command, when executed in the directory of a `.kicad_pcb` file, reads the PCB and generates three files:
- A `.zip` file containing the gerbers that can uploaded on the order page
- The `-cpl.csv` and `-BOM.csv` files that can be uploaded on the assembly page

If only basic parts are used, and if naming conventions are adhered to, then no
additional user input should be required. If extended components are used, then
the user will need to provide the LCSC part number for each extended component
not already associated. Again, because the association is global, these
associations only need to be provided once.

JCAD doesn't specify component rotations because these are corrected during DFM
review. The silkscreen should always allow the reviewer to correct component
rotations and the rotations given in the CPL file should onlybe seen as 
advisory.

## Configuring KiCad

A major advantage of JCAD is that to work with it, KiCad requires little or no
configuration beyond what is usually required. However, it's a good idea to
create custom symbol and footprint libraries to contain the parts commonly used
when manufacturing boards with JLCPCB. 

For basic resistors, capacitors, and inductors, the standard symbols should
be used and the value should be modified as specified in the naming convention,
rather than creating separate symbols for each value. In all cases, creating a
custom footprint should be a last resort used only when a builtin footprint
is not available.

Basic components should be used where possible and the resistor calculators
given below should be used to recalculate dividers to use values available
in basic parts: 

- https://damien.douxchamps.net/elec/resdiv/
- https://www.qsl.net/in3otd/parallr.html 


## Initial Configuration

JCAD currently requires a `go` compiler to build the program and requires
files that are no longer available on the JLCPCB website. The program is 
being modified to include the basic parts in the repository, but the
completion date of this work is unknown. 