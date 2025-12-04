from spiraltools import *
from quartic import *
import json
import math

def modify_timestamps( json_file_contents ):
    # time is important for the first few days where people are paying attention.
    # But time is NOT guaranteed to go up for successive blocks, and in fact DOESN'T!
    # mediantime is however guaranteed to go up. But it's the median of the last 11 blocks.
    # We will use time, and then call nudge_time_stamps() to enforce a minimum (positive!)
    # gap between adjacent blocks.
    # We've done away with the clunky ramping from time to mediantime (since time should be more
    # representative, though very occasionally quirky, for most blocks).
    prev_time = json_file_contents["Blocks"][0]["Time"]
    for block in json_file_contents["Blocks"]:
        time = block["Time"]
        median_time = block["MedianTime"]
        modified_time = max(prev_time + 1, time)    # Occasionally bunch up blocks, with one second interval between
        # Remove the old times, to be safe and reduce memory usage
        del block["Time"]
        del block["MedianTime"]
        block["modified_time"] = modified_time
        prev_time = modified_time

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

def towerMain():

    filename = "Input\\acidblocks3.json"
    print( "Read source data file", filename, "..." )
    fi1 = open(filename)
    json_file = json.load(fi1)
    fi1.close()

    print("Modify timestamps...")
    modify_timestamps(json_file)

    print( "Read json and create Instances with length, width, thickness, colour..." )
    instances = []
    for b, blockJson in enumerate(json_file["Blocks"]):
        size_bytes = blockJson["SizeBytes"]
        if size_bytes >= 16 * 16 * 16:
            length = math.pow(size_bytes, 1/3.0)
            width = math.pow(size_bytes, 1/3.0)
            thickness = math.pow(size_bytes, 1/3.0)
        elif size_bytes > 16 * 16:
            width = 16
            thickness = 16
            length = size_bytes / (16 * 16)
        else:
            length = 1
            width = 16
            thickness = size_bytes / 16
        red = blockJson["ColourByte0"] / 255.0
        green = blockJson["ColourByte1"] / 255.0
        blue = blockJson["ColourByte2"] / 255.0

        instances.append(Block(length, width, thickness, red, green, blue))
        instances[b]["modified_time"] = blockJson["modified_time"]

    print("Identify bunches (similar to days) of blocks...")
    # Bunches are typically days, but are also delineated where big gaps occur
    # Bunches can even be just one block long - eg the genesis block
    bunch_gap_limit = 60 * 60   # One hour
    instance_count = len(instances)
    first_block_of_bunch = 0
    prev_day = 0
    for i, instance in enumerate(instances):
        if i == 0:
            # Genesis block is the start and end of a bunch
            instances[0]["first_block_of_bunch"] = 0
            instances[0]["last_block_of_bunch"] = 0
            first_block_of_bunch = 0
        else:
            day = int(instance["modified_time"] / (24 * 60 * 60))      # Note that genesis block is not at day 0
            new_day = (day != prev_day)
            gap = instance["modified_time"] - instances[i-1]["modified_time"]
            new_bunch = (gap > bunch_gap_limit)
            if new_day or new_bunch or i == instance_count - 1:
                # Go back and fill in ["first_block_of_bunch"] and ["last_block_of_bunch"] for all blocks in the previous bunch
                for j in range(first_block_of_bunch, i):
                    instances[j]["first_block_of_bunch"] = first_block_of_bunch
                    instances[j]["last_block_of_bunch"] = i - 1
                    # If i is the last block of the whole blockchain, also fill in i
                    if i == instance_count - 1:
                        instances[i]["first_block_of_bunch"] = first_block_of_bunch
                        instances[i]["last_block_of_bunch"] = i
                first_block_of_bunch = i
            prev_day = day

    print("Calculate day_angle and day_radius_raw...")
    for i, instance in enumerate(instances):
        if instance["first_block_of_bunch"] == i:
            length_for_bunch = instance.length / 2.0
        elif instance["last_block_of_bunch"] == i:
            length_for_bunch += instance.length / 2.0
        else:
            length_for_bunch += instance.length
        if instance["last_block_of_bunch"] == i:
            # Deal with each block in the bunch

            # Special case for a bunch that is a single block.
            # Avoids a divide by zero
            if instance["first_block_of_bunch"] == instance["last_block_of_bunch"]:
                timestamp = instance["modified_time"]
                second_of_day = timestamp % (24 * 60 * 60)
                # Artistic license! Make the genesis vertical so you can read the words!
                if i == 0:
                    second_of_day = 12 * 60 * 60
                day_angle = 360.0 * second_of_day / (24 * 60 * 60)
                # Add 270 degrees and negate to make day spiral clockwise with midnight at the top
                day_angle = (270.0 + 360.0 - day_angle) % 360.0
                instances[i]["day_angle"] = day_angle

                instances[i]["day_radius_raw"] = 0.0    # This is no use, but will be smoothed out by the quartic dips

            else:
                # Go back and calculate for the whole bunch
                first_timestamp_of_bunch = instances[instance["first_block_of_bunch"]]["modified_time"]
                second_of_day = first_timestamp_of_bunch % (24 * 60 * 60)
                fraction_of_revolution_start =  second_of_day / (24 * 60 * 60)
                first_angle_of_bunch = 360.0 * fraction_of_revolution_start

                last_timestamp_of_bunch = instances[i]["modified_time"]
                second_of_day = last_timestamp_of_bunch % (24 * 60 * 60)
                fraction_of_revolution_end =  second_of_day / (24 * 60 * 60)
                last_angle_of_bunch = 360.0 * fraction_of_revolution_end

                time_for_bunch = last_timestamp_of_bunch - first_timestamp_of_bunch

                length_so_far = 0.0
                for j in range(instance["first_block_of_bunch"], i + 1):
                    # First calculate fraction_of_bunch_length (based on material lengths)
                    length = instances[j].length
                    length_so_far += length / 2.0
                    fraction_of_bunch_length = length_so_far / length_for_bunch
                    length_so_far += length / 2.0

                    # Second calculate fraction_of_rev_time (based on time gaps)
                    timestamp = instances[j]["modified_time"]
                    fraction_of_bunch_time = (timestamp - first_timestamp_of_bunch) / time_for_bunch

                    # Third, combine the two based on a time:space ratio
                    time_weighting = 1.0  # Equal weighting to time (gaps) and space (material)
                    space_weighting = 1.0
                    fraction_of_bunch = (fraction_of_bunch_time * time_weighting + fraction_of_bunch_length * space_weighting) / (time_weighting + space_weighting)

                    day_angle = first_angle_of_bunch + (last_angle_of_bunch - first_angle_of_bunch) * fraction_of_bunch

                    # Add 90 and negate to make day spiral clockwise with midnight at the top
                    day_angle = (270.0 + 360.0 - day_angle) % 360.0
                    instances[j]["day_angle"] = day_angle

                    # We have a length_for_bunch which we can multiply by (time_weighting + space_weighting)
                    # to give a partial-circumference for the bunch. And we have a range of angles we know it applies
                    # across. Thus we can calculate a day radius.
                    partial_circumference = length_for_bunch * (time_weighting + space_weighting)
                    partial_revolution = (last_angle_of_bunch - first_angle_of_bunch) / 360.0
                    full_circumference = (partial_circumference / partial_revolution)
                    day_radius = full_circumference / (2.0 * math.pi)
                    instances[j]["day_radius_raw"] = day_radius

    print("Find the maxima of day_radius within neighbourhoods of a certain size...")
    maxima_indices = find_maxima_indices(instances, "day_radius_raw", 144)

    print("Smooth between maxima of day_radius using quartic dips...")
    label_quartic_dips(instances, maxima_indices, "modified_time", "day_radius_raw", "day_radius")

    filename = 'quartics.csv'
    print("Outputting quartic dips debug csv to", filename, "...")
    with open(filename, 'w') as f:
        for index in range(0,100000):
            if index in maxima_indices:
                # Put spikes in third column to indicate maxima
                print(index, ",", instances[index]["modified_time"] - 1230768000, ",", instances[index]["day_radius_raw"], ",", instances[index]["day_radius"], ",", instances[index]["day_radius_raw"]/2, file=f)
            else:
                print(index, ",", instances[index]["modified_time"] - 1230768000, ",", instances[index]["day_radius_raw"], ",", instances[index]["day_radius"], ",", 0.0, file=f)

    print("Introduce transforms to each instance...")
    for i, instance in enumerate(instances):
        # Half block thickness so inside cylinder of dayLoop is smooth
        half_thickness = instance.thickness / 2
        instance.introducedTransforms.append(TranslateX(half_thickness))

        # Give day a radius
        day_radius = instance["day_radius"]
        instance.introducedTransforms.append(TranslateX(day_radius))

        # Rotation for elements of day loop
        instance.introducedTransforms.append(RotateY(instance["day_angle"]))

        # Give year a radius
        year_radius = 10000    # stab in the dark for now
        instance.introducedTransforms.append(TranslateX(year_radius))

        # Rotation for elements of year loop
        timestamp = instance["modified_time"]
        first_jan_2009_midnight = 1230768000
        year_second = (timestamp - first_jan_2009_midnight) % (365.25 * 24 * 60 * 60)
        year_angle = 360.0 * year_second / (365.25 * 24 * 60 * 60)
        instance.introducedTransforms.append(RotateZ(year_angle))

    print("'Render'...")
    renderer = []                   # Renderer can merely be an array to append to
    for instance in instances:
        instance.render(renderer)

    filename = "Output\\renderspec.json"
    print( "Saving (this will take a while) to", filename, "..." )
    fo = open(filename, 'w')
    json.dump(renderer, fo, default=vars, indent=2)

towerMain()