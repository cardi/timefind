#!/usr/bin/python

import argparse
import csv
import os

parser = argparse.ArgumentParser()
parser.add_argument('files', metavar='file', nargs="+", type=file,
                    help='list of index files to parse')
parser.add_argument('-d', '--directory', metavar="DIR", default="./newindex",
                    help="Directory to create output subdirectories in and new split file")
parser.add_argument('-f', '--filename', metavar="NAME", default="pcap.csv",
                    help="The name of the index file in the subdirs to create")
parser.add_argument('-n', '--number', metavar="NUM", default=1000, type=int,
                    help="Number of entries to include in each new subdirectory")
args = parser.parse_args()

os.mkdir(args.directory)
masteroutput = args.directory + "/" + args.filename
masteroutputf = open(masteroutput, "w")
masteroutputcsv = csv.writer(masteroutputf)


for fileh in args.files:
    reader = csv.reader(fileh)
    count=0
    newbeginning = 0

    for row in reader:
        filename = row[0]
        begin = row[1]
        end = row[2]

        if count % args.number == 0:
            # write the results of the last collection
            if count > 0:
                masteroutputcsv.writerow([newsubdir, newbeginning, end, 9999])

            # create the new direct
            newsubdir = ("%04d" % (count / args.number))
            newdir = args.directory + "/" + newsubdir + "/"
            os.mkdir(newdir)
            outf = open(newdir + args.filename, "w")
            csvwriter=csv.writer(outf)

            # remember the starting timestamp
            newbeginning = begin

        csvwriter.writerow(row)
        count = count + 1

    # save the final row
    masteroutputcsv.writerow([newsubdir, newbeginning, end, 9999])
