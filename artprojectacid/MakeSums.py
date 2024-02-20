import json
import math

def MakeSums(inputFileIndices, inputFileUnsquished, outputFile):
    fi1 = open(inputFileIndices)
    indices = json.load(fi1)

    fi2 = open(inputFileUnsquished)
    unsquished = json.load(fi2)

    sums = {}

    # sums.blocks
    sums["blocks"] = len(indices["blocks"])

    # sums.ByYear.firstDay[]
    # sums.ByYear.lastDay[]
    # sums.ByDay.firstBlock[]
    # sums.ByDay.lastBlock[]
    sums["ByYear"] = {}
    sums["ByYear"]["firstDay"] = []
    sums["ByYear"]["lastDay"] = []
    sums["ByDay"] = {}
    sums["ByDay"]["firstBlock"] = []
    sums["ByDay"]["lastBlock"] = []
    prevBlockDay = -1
    prevBlockYear = -1
    prevBlockNum = -1
    for block in indices["blocks"]:
        day = block["day"]
        year = block["year"]
        if day > prevBlockDay:
            # First block of a day
            if prevBlockDay != -1:
                sums["ByDay"]["lastBlock"].append(prevBlockNum)
            sums["ByDay"]["firstBlock"].append(prevBlockNum + 1)
            if year > prevBlockYear:
                # First day/block of a year
                if prevBlockYear != -1:
                    sums["ByYear"]["lastDay"].append(prevBlockDay)
                sums["ByYear"]["firstDay"].append(day)
        prevBlockDay = day
        prevBlockYear = year
        prevBlockNum = prevBlockNum + 1
    sums["ByDay"]["lastBlock"].append(prevBlockNum)
    sums["ByYear"]["lastDay"].append(prevBlockDay)
    print("Last Block = ", prevBlockNum)
    print("Last day = ", prevBlockDay)

    # sums.ByDay.Md[]
    # sums.ByDay.Hd[] (same as Md! ToDo)
    # sums.ByDay.Nd[]
    # sums.ByYear.Jy[]
    # sums.dayRadiusGuide
    # sums.yearRadiusGuide
    # sums.centuryRadiusGuide
    sums["ByDay"]["Md"] = []
    sums["ByDay"]["Hd"] = []
    sums["ByDay"]["Nd"] = []
    sums["ByDay"]["SumLb"] = []
    sums["ByDay"]["Ld"] = []
    sums["ByDay"]["MaxTb"] = []
    sums["ByDay"]["MaxWb"] = []
    sums["ByYear"]["Jy"] = []
    sums["ByYear"]["SumLd"] = []
    totalLength = 0
    totalWidth = 0
    totalDayMaxWb = 0
    countDays = 0
    for year, firstDay in enumerate(sums["ByYear"]["firstDay"]):
        lastDay = sums["ByYear"]["lastDay"][year]
        Jy = 0
        SumLd = 0
        for day in range(firstDay, lastDay+1):
            print("Day=", day)
            Md = 0
            Jd = 0
            Nd = 0
            SumLb = 0
            Ld = 0
            MaxTb = 0
            MaxWb = 0
            firstBlock = sums["ByDay"]["firstBlock"][day]
            lastBlock = sums["ByDay"]["lastBlock"][day]
            for block in range(firstBlock, lastBlock+1):
                volume = unsquished["blocks"][block]["volume"]
                Lb = volume["length"]       # Positive
                Wb = volume["width"]        # Positive
                Tb = volume["thickness"]    # Positive
                totalLength += Lb
                totalWidth += Wb
                SumLb += Lb
                Ld = max(Ld, Wb)
                MaxTb = max(MaxTb, Tb)
                MaxWb = max(MaxWb, Wb)
                # Allow 0/0=1 to avoid division by zero
                if Tb==0 and Wb==0:
                    Fb = 1 + Lb             # Positive
                    Gb = 1 + Lb             # Positive
                else:
                    Fb = 1 + Lb * Tb / Wb   # Positive
                    Gb = 1 + Lb * Wb / Tb   # Positive
                Hb = Lb / Fb / Gb           # Positive
                Jb = Wb * Fb                # Positive
                Nb = Tb / Fb / Gb           # Positive

                Md += Hb
                Jd = max(Jd, Jb)
                Nd = max(Nd, Nb)
            sums["ByDay"]["Md"].append(Md)
            sums["ByDay"]["Hd"].append(Md)  # Oops Hd is actually same as Md!
            sums["ByDay"]["Nd"].append(Nd)
            sums["ByDay"]["SumLb"].append(SumLb)
            sums["ByDay"]["MaxTb"].append(MaxTb)
            sums["ByDay"]["MaxWb"].append(MaxWb)
            totalDayMaxWb += MaxWb
            Jy += Jd
            SumLd += Ld
            countDays += 1
        sums["ByYear"]["Jy"].append(Jy)
        sums["ByYear"]["SumLd"].append(SumLd)

    averageLength = totalLength / sums["blocks"]
    averageWidth = totalWidth / sums["blocks"]
    averageMaxWb = totalDayMaxWb / countDays
    sums["dayRadiusGuide"] = averageLength * 144 / 2 / math.pi
    sums["yearRadiusGuide"] = averageMaxWb * 365 / 2 / math.pi
    sums["centuryRadiusGuide"] = sums["dayRadiusGuide"] * 100 / 2 / math.pi

    fo = open(outputFile, 'w')
    json.dump(sums,fo,indent=2)

    print("MakeSums: Finished")
