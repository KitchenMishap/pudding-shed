import MakeCuboids
import MakeIndices
import MakeVolumes
import MakeSums
import shutil

print("Starting...")

# Process the input files
MakeCuboids.MakeCuboids('Input\\acidblocks.json', 'Intermediate\\cuboids.json')
MakeIndices.MakeIndices('Input\\acidblocks.json', 'Intermediate\\indices.json')

# Process the intermediate files
MakeVolumes.MakeVolumes('Intermediate\\cuboids.json', 'Intermediate\\unsquished.json')
MakeSums.MakeSums('Intermediate\\indices.json', 'Intermediate\\unsquished.json', 'Intermediate\\sums.json')

# Gather the output files
shutil.copyfile('Intermediate\\unsquished.json', 'Output\\unsquished.json')
shutil.copyfile('Intermediate\\cuboids.json', 'Output\\cuboids.json')

# Gather files to input to SpiralOpt
shutil.copyfile('Input\\config.json', 'SpiralOptInput\\config.json')
shutil.copyfile('Input\\spacings.json', 'SpiralOptInput\\spacings.json')
shutil.copyfile('Intermediate\\unsquished.json', 'SpiralOptInput\\unsquished.json')
shutil.copyfile('Intermediate\\sums.json', 'SpiralOptInput\\sums.json')

# You still need to use SpiralOpt to generate coords.json from the above
