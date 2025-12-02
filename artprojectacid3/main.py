from spiraltools import *
from quartic import *
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
    # (and even series of blocks) with the same timestamp. Here we impose a minimum delta time between blocks
    # - some blocks may be moved forward in time to impose this, cumulatively, until there is space to catch up
    # in future blocks.

    blocks = json_file["Blocks"]
    time = blocks[0]["SpiralTime"]
    cumulative_time = time

    for b, block in enumerate(blocks):
        if b >=1:
            time = block["SpiralTime"]
            cumulative_time += min_delta_time
            if cumulative_time <= time:
                cumulative_time = time
            elif b < 1000:
                print( "Block ", b, ": Cumulative time ahead by ", cumulative_time - time)
            block["SpiralTime"] = cumulative_time

def is_maximum(instances, index, name_attr, half_period):
    value = instances[index][name_attr]
    for i in range(1, half_period + 1):
        matched_left = False
        matched_right = False
        left = instances[index - i][name_attr]
        if left > value:
            return False    # Something nearby is bigger. Definitely return false
        right = instances[index + i][name_attr]
        if right > value:
            return False    # Something nearby is bigger. Definitely return false
        if left == value:
            matched_left = True     # Something nearby on the left is equal. Special treatment!
            if matched_left and matched_right:
                return False        # definitely return false if matched on both sides nearby
        if right == value:
            matched_right = True    # Something nearby on the right is equal. Special treatmeent!
            if matched_left and matched_right:
                return False        # definitely return false if matched on both sides nearby
    return True     # Nothing nearby is bigger. And (subtly special case) if matched, it is only matched on ONE side nearby

def find_maxima_indices(instances, name_attr, half_period):
    length = len(instances)
    results = []
    if length < half_period + half_period + 1:
        return results
    for i in range(half_period, length - half_period - 1):
        if is_maximum(instances, i, name_attr, half_period):
            results.append(i)
    return results

def label_quartic_dips(instances, maxima_indices, name_time, name_source, name_target):
    index_index = 1     # For now, maxima_indices SEEM to come in pairs 1966, 1967, 4098, 4099, 7527, 7528 .. 821790, 821791
                        # NOT ALWAYS TRUE
    while index_index < len(maxima_indices) - 1:
        period_start_index = maxima_indices[index_index]
        period_end_index = maxima_indices[index_index + 1]
        points = []
        for index in range(period_start_index, period_end_index + 1):
            x = instances[index][name_time]
            y = instances[index][name_source]
            points.append([x,y])
        mean_x_val = mean_x(points)
        mean_y_val = mean_y(points)
        C_offset = make_quartic_dip_offset(points, 10, mean_x_val, mean_y_val)
        for index in range(period_start_index, period_end_index + 1):
            x = instances[index][name_time]
            y = quartic_curve_offset(x, C_offset, mean_x_val, mean_y_val)
            instances[index][name_target] = y
        while maxima_indices[index_index + 1] == maxima_indices[index_index]:
            index_index += 1
        index_index += 1

    # Extend the last maxima, flat to the end of the sequence of instances
    flat_value = y
    for index in range(period_end_index + 1, len(instances)):
        instances[index][name_target] = flat_value

    # Extend the first maxima, flat back to the start of the sequence of instances
    index_index = 1
    period_start_index = maxima_indices[index_index]
    flat_value = instances[period_start_index][name_target]
    for index in range(period_start_index):
        instances[index][name_target] = flat_value

# The length-per-day of a block is the sum length of blocks between 12h before and 12h after that block
# The active_hours_per_day is the number of whole hours that contain blocks over this 24h period
def label_length_per_day(instances):
    for i, instance in enumerate(instances):
        active_hours = {}
        timestamp = instance["SpiralTime"]
        time_start = timestamp
        time_end = timestamp
        length = 0

        # Search backwards
        j = i-1
        while j >= 0 and instances[j]["SpiralTime"] > timestamp - 12 * 60 * 60:
            length += instances[j].length
            time_start = instances[j]["SpiralTime"]
            hour = int(time_start / (60*60))
            active_hours[hour] = True
            j -= 1

        # Search forwards
        j = i
        while j < len(instances) and instances[j]["SpiralTime"] < timestamp + 12 * 60 * 60:
            length += instances[j].length
            time_end = instances[j]["SpiralTime"]
            hour = int(time_end / (60 * 60))
            active_hours[hour] = True
            j += 1

        instance["length_per_day"] = length
        instance["active_hours_per_day"] = len(active_hours)

def towerMain():

    print( "Opening source data file..." )
    fi1 = open("Input\\acidblocks3.json")
    json_file = json.load(fi1)
    fi1.close()

    print("Generate SpiralTime...")
    generate_spiral_time(json_file)

    print("Nudge Timestamps...")
    nudge_time_stamps(json_file, 60)

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
        instances[b]["SpiralTime"] = blockJson["SpiralTime"]

    print("Pass 1.1, label up length and active hours per day")
    label_length_per_day(instances)

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
        r_limit_before = r_limit_array[b-1]         # For block 1, look in gap 0
        r_limit_after = r_limit_array[b]            # for block 1, look in gap 1
        instances[b]["r_min_day"] = max(r_limit_before, r_limit_after)

    print("Fourth pass, find the maxima within neighbourhoods of a certain size")
    maxima_indices = find_maxima_indices(instances, "r_min_day", 144)
    print(maxima_indices)

    print("Fifth pass, label quartic dips - COMMENTED OUT")
    #label_quartic_dips(instances, maxima_indices, "SpiralTime", "r_min_day", "r_day")

    print("Pass 5.1, day radius based on length_per_day")
    for instance in instances:
        active_hours_per_day = instance["active_hours_per_day"]
        if active_hours_per_day == 0:
            active_hours_per_day = 1    # Avoid divide by zero for genesis block
        instance["r_day"] = (instance["length_per_day"] / (2.0 * math.pi)) * (24 / active_hours_per_day)

    with open('quartics.csv', 'w') as f:
        for index in range(0,100000):
            if index in maxima_indices:
                # Put spikes in third column to indicate maxima
                print(index, ",", instances[index]["SpiralTime"] - 1230768000, ",", instances[index]["r_min_day"], ",", instances[index]["r_day"], ",", instances[index]["r_day"]/2, file=f)
            else:
                print(index, ",", instances[index]["SpiralTime"] - 1230768000, ",", instances[index]["r_min_day"], ",", instances[index]["r_day"], ",", 0.0, file=f)

    print("Sixth pass, introduce transforms...")
    for i, instance in enumerate(instances):
        blockJson = json_file["Blocks"][i]

        # Half block thickness so inside cylinder of dayLoop is smooth
        half_thickness = instance.thickness / 2
        instance.introducedTransforms.append(TranslateX(half_thickness))

        # Give day a radius
        day_radius = instance["r_day"]
        instance.introducedTransforms.append(TranslateX(day_radius))

        # Rotation for elements of day loop
        timestamp = instance["SpiralTime"]
        first_jan_2009_midnight = 1230768000
        day_second = (timestamp - first_jan_2009_midnight) % (24 * 60 * 60)
        day_angle = 360.0 * day_second / (24 * 60 * 60)
        instance.introducedTransforms.append(RotateY(day_angle))

        # Give year a radius
        year_radius = 10000    # stab in the dark for now
        instance.introducedTransforms.append(TranslateX(year_radius))

        # Rotation for elements of year loop
        year_second = (timestamp - first_jan_2009_midnight) % (365.25 * 24 * 60 * 60)
        year_angle = 360.0 * year_second / (365.25 * 24 * 60 * 60)
        instance.introducedTransforms.append(RotateZ(year_angle))

    print("Seventh pass, render")
    renderer = []                   # Renderer can merely be an array to append to
    for instance in instances:
        instance.render(renderer)

    print( "Save..." )
    fo = open("Output\\renderspec.json", 'w')
    json.dump(renderer, fo, default=vars, indent=2)

towerMain()