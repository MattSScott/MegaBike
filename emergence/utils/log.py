import logging

# initialise logging to a file named by the current hostname
log_path = f"EmergenceCalc.log"
logging.basicConfig(filename = log_path, filemode = 'a', level = logging.INFO,
    format = '%(asctime)s.%(msecs)03d %(message)s', datefmt = '%Y%m%d %H:%M:%S')
