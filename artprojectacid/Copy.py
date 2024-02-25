import shutil

print("Copying...")

# Gather the final output file
shutil.copyfile('SpiralOptOutput\\coords.json', 'Output\\coords.json')

# Copy output files to spiralviz2
shutil.copyfile('Output\\coords.json', 'C:\\Program Files\\Epic Games\\UE_5.2\\Engine\\Binaries\\Win64\\coords.json')
shutil.copyfile('Output\\cuboids.json', 'C:\\Program Files\\Epic Games\\UE_5.2\\Engine\\Binaries\\Win64\\cuboids.json')
shutil.copyfile('Output\\unsquished.json', 'C:\\Program Files\\Epic Games\\UE_5.2\\Engine\\Binaries\\Win64\\unsquished.json')

# Now you need to run SpiralViz2
