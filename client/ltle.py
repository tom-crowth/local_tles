#!/usr/env/bin python

import glob
import os
import shutil
import subprocess
from skyfield.api import EarthSatellite, load
import json
from datetime import datetime, timedelta
import numpy as np

ts = load.timescale(builtin=True)

go_prog = shutil.which("local_tles")

MAX_DAYS = 7
DATE_FORMAT = "%Y-%m-%d"


class LocalTLE:
    def __init__(self):
        self.parse_config_file()

    def download_tles_for_date(self, date_str):
        # check for already running local_tles here
        # otherwise run a one-off
        print("Need to check for running program here in future...")

        # check if within allowed date range from config
        if not self.config.check_date(date_str):
            print("TLEs will not be downloaded")
            return

        # check if it already exists
        if self.config.check_date_file(date_str):
            print(f"TLE file for date {date_str} already exists. If you want to re-download, run with force=True (not yet implemented)")
            return
            
        cmd = [go_prog, date_str]
        subprocess.run(cmd)

    def download_tles_for_date_range(self, date_range):
        date_strs = self.get_date_strings_for_range(date_range)
        if date_strs is None:
            return

        # check all first
        results = self.config.check_date_file_list(date_strs)
        if False in results:
            num_downloads = (np.array(results) == False).size
            print(f"There are {num_downloads} downloads required for this date range...")
        else:
            print(f"No downloads required for this date range. All files already exist.")
            return

        for date in date_strs:
            print(f"DEBUG: running for {date}")
            if self.config.check_date_file(date):
                print(f"Data for {date} already exists. Continuing")
                continue
            self.download_tles_for_date(date)

    def get_date_strings_for_range(self, date_range):
        split_range = date_range.split('--') # as per SpaceTrack API
        if len(split_range) != 2:
            print(f"Invalid date range string. {date_range}")
        first_date, last_date = split_range
        first_date = datetime.strptime(first_date, DATE_FORMAT)
        last_date = datetime.strptime(last_date, DATE_FORMAT)
        diff = (last_date - first_date).days
        if diff > MAX_DAYS:
            print(f"Date range {date_range} exceeds maximum number of days {MAX_DAYS}")
            return
        date_strs = []
        for i in range(diff+1):
            current = first_date + timedelta(days=i)
            current_str = current.strftime(DATE_FORMAT)
            date_strs.append(current_str)
        return date_strs

    def get_latest_data(self):
        latest = self.config.last_date_file()
        data = process_file(latest)
        return data

    def get_data_for_date(self, date_str):
        if not self.config.check_date_file(date_str):
            print(f"Error: date_str specified: {date_str} does not have a data file. You will need to get TLEs for this date with download_tles_for_date()")
            return
        print(f"Reading data for {date_str}...")
        data = process_file(self.config.get_date_files(date_str))
        return data

    def get_data_for_date_range(self, date_range, filter=False):
        data = []
        date_strs = self.get_date_strings_for_range(date_range)
        if date_strs is None:
            return
        # check all before reading any
        for date in date_strs:
            if not self.config.check_date_file(date):
                print(f"Date file {date} does not exist. Exiting")
                return

        for date in date_strs:
            data.extend(self.get_data_for_date(date))
        return data
        

    def parse_config_file(self):
        ltle_config = os.getenv("LTLE_CONFIG")
        if ltle_config is None:
            print(f"Config file could not be found. Exiting")
            return
        config_file = os.path.join(ltle_config, "ltle.config")
        self.config = LocalTLEConfig(config_file)




class LocalTLEConfig:
    def __init__(self, config_file):
        if not os.path.exists(config_file):
            print(f"Error: specified config file {config_file} does not exist.")
            return
        with open(config_file, 'r') as f:
            self.raw_lines = []
            for line in f.readlines():
                line = line.rstrip('\n')
                self.raw_lines.append(line)
                if line == "":
                    continue
                items = line.split(" ")
                name, val = items
                if name == "READ_RATE":
                    self.read_rate = int(val)
                elif name == "PORT":
                    self.port = int(val)
                elif name == "FIRST_DATE":
                    self.first_date = val
                elif name == "DATASTORE":
                    self.datastore = val
                elif name == "DATABASE":
                    self.database = val
                elif name == "ACCESS_ST":
                    self.access_flag = True if val else False
                else:
                    print(f"Error: something wrong with the config file. Received key {name}")

        # carry out checks
        if not os.path.exists(self.datastore):
            print(f"Specified datastore location {self.datastore} does not exist.")
        if not os.path.exists(self.database):
            print(f"Specified database location {self.database} does not exist.")

    def check_date(self, date_str):
        if self.first_date > date_str:
            # should be able to just carry out alphabetic order check on date_str
            print(f"Date string {date_str} is before first date {self.first_date}")
            return False
        else:
            return True

    def get_date_files(self, date_str):
        filename_glob = f"{date_str}*.data"
        files = sorted(glob.glob(os.path.join(self.datastore, filename_glob)))
        if len(files) > 1:
            return files
        elif len(files) == 1:
            return files[0]
        else:
            return False

    def check_date_file(self, date_str):
        return not (self.get_date_files(date_str) == False)

    def check_date_file_list(self, date_strs):
        results = []
        for date_str in date_strs:
            results.append(self.check_date_file(date_str))
        return results

    def last_date_file(self):
        files = sorted(glob.glob(os.path.join(self.datastore, '*')))
        return files[-1]



def get_sat(sat_json):
    sat = EarthSatellite.from_omm(ts, sat_json)
    return sat

def process_satlist(tle_list_json, filters=None):
    # assumes metadata is also included
    # ignores it ('request_metadata')
    sat_list = []
    if type(filters) == dict:
        filters = [filters]
    # TODO: check for data key first
    for sat_json in tle_list_json['data']:
        if filters is not None:
            valid = True
            for filter in filters:
                field = filter['field']
                operator = filter['operator']
                val_type = filter['type']
                limit = filter['value']
                satval = sat_json[field]
                if val_type == 'numerical':
                    limit = float(limit)
                    satval = float(satval)
                if operator == 'lte' and satval > limit:
                    valid = False
                    break
                elif operator == 'lt' and satval >= limit:
                    valid = False
                    break
                elif operator == 'gte' and satval < limit:
                    valid = False
                    break
                elif operator == 'gt' and satval <= limit:
                    valid = False
                    break
                elif operator == 'eq' and satval != limit:
                    valid = False
                    break
                elif operator == 'neq' and satval == limit:
                    valid = False
                    break
            if valid:
                sat_list.append(get_sat(sat_json))
        else:
            sat_list.append(get_sat(sat_json))
    return sat_list

def raw_data(data_file):
    if type(data_file) != list:
        data_file = [data_file]
    sat_list = []
    for file in data_file:
        with open(file, 'r') as f:
            data = json.load(f)
            sat_list.extend(data)

    return sat_list


def process_file(data_file, filters=None):
    if type(data_file) != list:
        data_file = [data_file]
    sat_list = []
    for file in data_file:
        with open(file, 'r') as f:
            data = json.load(f)
            sat_list.extend(process_satlist(data, filters=filters))

    return sat_list

def add_filter(field, value, operator, val_type='numerical'):
    filter = {
        'field': field,
        'value': value,
        'operator': operator,
        'type': val_type,

    }
    return filter

def filter_sats_by_time(sats, time):
    time_diffs = [None] * len(sats)

    for i in range(len(sats)):
        satellite = sats[i]
        time_diffs[i] = abs(satellite.epoch - time)
    
    sat_times = list(zip(sats, time_diffs))
    sat_times.sort(key=lambda x: x[0].name)

    # remove duplicates based on time difference from tle to request
    i = 0
    end = len(sat_times) - 1
    while i < end:
        if sat_times[i][0].name == sat_times[i+1][0].name:
            if sat_times[i][1] > sat_times[i+1][1]:
                sat_times.pop(i)
                #i += 1
            else:
                sat_times.pop(i+1)
            end -= 1
        else:
            i += 1
    if len(sat_times) == 1:
        return sat_times[0][0]
    return [i for i, _ in sat_times]

# TODO: add processing to select *closest* TLE for given object (to a specific time), rather than using all of them
