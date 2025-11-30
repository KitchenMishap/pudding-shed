from spiraltools import *
import json
import math

def generate_spiral_time( json_file ):
    # time is important for the first few days where people are paying attention.
    # But time is NOT guaranteed to go up for successive blocks, and in fact DOESN'T!
    # mediantime is however guaranteed to go up. But it's the median of the last 11 blocks.
    # So we introduce "spiraltime", which starts as "time" but transitions gradually to "mediantime"
    start_ramp_block = 2016     # Arbitrarily coincide the transition ramp with the second difficulty epoch
    ramp_blocks_length = 2016.0
    blocks = len(json_file["Blocks"])
    for b in range(blocks):
        if b < start_ramp_block:
            ramp_param = 0.0                # entirely "time"
        elif b - start_ramp_block < ramp_blocks_length:
            ramp_param = (b - start_ramp_block) / ramp_blocks_length
        else:
            ramp_param = 1.0                # entirely "mediantime"
        time = json_file["Blocks"][b]["Time"]
        median_time = json_file["Blocks"][b]["MedianTime"]
        spiral_time = ramp_param * median_time + (1.0 - ramp_param) * time
        # Remove the old times, to be safe
        del json_file["Blocks"][b]["Time"]
        del json_file["Blocks"][b]["MedianTime"]
        json_file["Blocks"][b]["SpiralTime"] = spiral_time

def nudge_time_stamps( json_file, min_delta_time ):
    # Even when using spiral_time (see above, which is usually median_time), we find may adjacent blocks
    # (and even series of blocks) with the same timestamp. Here we detect sequences with small delta_times between
    # them, and try to spread them out a bit, using the surrounding blocks to provide a scale
    failures = 0

    blocks = len(json_file["Blocks"])
    b = 1
    delta_time = json_file["Blocks"][b]["SpiralTime"] - json_file["Blocks"][b-1]["SpiralTime"]

    while True:
        # Search for the start of the next small delta_time sequence
        while delta_time >= min_delta_time:
            b = b + 1
            if b >= blocks:
                print( failures, " failures")
                return
            delta_time = json_file["Blocks"][b]["SpiralTime"] - json_file["Blocks"][b-1]["SpiralTime"]

        # Found a small delta_time between b-1 and b
        first_short_block = b

        # Search for the end of this small delta_time sequence
        while delta_time < min_delta_time and b < blocks - 1:
            b = b + 1
            delta_time = json_file["Blocks"][b]["SpiralTime"] - json_file["Blocks"][b-1]["SpiralTime"]

        # Rare edge case: the sequence of short blocks is at the end of the chain
        if b == blocks - 1:
            time_stamp = json_file["Blocks"][first_short_block - 1]["SpiralTime"]
            for b in range(first_short_block, blocks):
                time_stamp += min_delta_time
                json_file["Blocks"][b]["SpiralTime"] = time_stamp
            print( failures, " failures")
            return

        last_short_block = b - 1
        before_time = json_file["Blocks"][first_short_block - 1]["SpiralTime"]
        after_time = json_file["Blocks"][last_short_block + 1]["SpiralTime"]
        count = last_short_block + 1 - first_short_block
        step = float(after_time - before_time) / (count + 2.0)
        #if count > 1:
        #    print( "Attempting to remedy ", count, " short blocks with step ", step)
        if step < min_delta_time:
            print( "Failed to remedy ", count, "short blocks at ", first_short_block, ", step ", step, " is less than limit ", min_delta_time)
            failures = failures + 1
        time_stamp = float(before_time)
        for bb in range(first_short_block, last_short_block + 1):
            time_stamp += step
            json_file["Blocks"][bb]["SpiralTime"] = time_stamp

def towerMain():

    print( "Opening source data file..." )
    fi1 = open("Input\\acidblocks3.json")
    json_file = json.load(fi1)
    fi1.close()

    print("Generate SpiralTime...")
    generate_spiral_time(json_file)

    print("Nudge Timestamps...")
    nudge_time_stamps(json_file, 2)

    print( "First pass: Import json and create block Instances with length, width, thickness" )
    instances = []
    for b, blockJson in enumerate(json_file["Blocks"]):
        sizeBytes = blockJson["SizeBytes"]
        if sizeBytes >= 16 * 16 * 16:
            length = math.pow(sizeBytes, 1/3.0)
            width = math.pow(sizeBytes, 1/3.0)
            thickness = math.pow(sizeBytes, 1/3.0)
        elif sizeBytes > 16 * 16:
            width = 16
            thickness = 16
            length = sizeBytes / (16 * 16)
        else:
            length = 1
            width = 16
            thickness = sizeBytes / 16
        red = blockJson["ColourByte0"] / 255.0
        green = blockJson["ColourByte1"] / 255.0
        blue = blockJson["ColourByte2"] / 255.0

        instances.append(Block(length, width, thickness, red, green, blue))

    print("Second pass, mark up the dayRadiusRLimit's of gaps between blocks")
    # In these sections, for things with r in the name, r refers to the day radius
    block_count = len(instances)
    gap_count = block_count - 1
    r_limit_array = []
    for g in range(gap_count):
        prev_block = g
        next_block = g + 1
        a = instances[prev_block].length / 2.0
        b = instances[next_block].length / 2.0
        delta_time = json_file["Blocks"][next_block]["SpiralTime"] - json_file["Blocks"][prev_block]["SpiralTime"]
        if delta_time < 1:
            print("Gap ", g, ": delta time < 1:", delta_time)
            delta_time = 1
        theta = 2.0 * math.pi * delta_time / (24.0 * 60.0 * 60.0)
        r_limit_array.append((a+b) / theta)

    print("Third pass, use rLimit's on gaps to determine rMin on neighbouring blocks")
    # End blocks have only one r_limit to make a judgement on
    instances[0]["r_min_day"] = r_limit_array[0]
    instances[block_count - 1]["r_min_day"] = r_limit_array[gap_count - 1]
    # All other blocks are the max of the r_limits on the gaps before/after
    for b in range(1, block_count-1):
        r_limit_before = r_limit_array[b-1]     # For block 1, look in gap 0
        r_limit_after = r_limit_array[b]        # for block 1, look in gap 1
        instances[b]["r_min_day"] = max(r_limit_before, r_limit_after)

    print("Second pass, introduce transforms...")
    for i, instance in enumerate(instances):
        blockJson = json_file["Blocks"][i]

        # Half block thickness so inside cylinder of dayLoop is smooth
        halfThickness = instance.thickness / 2
        instance.introducedTransforms.append(TranslateX(halfThickness))

        # Give day a radius
        #instance.introducedTransforms.append(TranslateX(dayRadius))

        # Rotation for elements of day loop
        #instance.introducedTransforms.append(RotateY(dayAngle))

        # Give year a radius
        #instance.introducedTransforms.append(TranslateX(yearRadius))

        # Rotation for elements of year loop
        #instance.introducedTransforms.append(RotateZ(yearAngle))

towerMain()