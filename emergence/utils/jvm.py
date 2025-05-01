"""
Tools for interfacing with Java classes using the JVM via jpype
"""
import jpype as jp
import logging
import numpy as np
import os
import sys
import typing

import emergence.utils.log

class JVM:
    """
    Singleton class for managing the JVM for calls to infodynamics.jar
    """
    INFODYNAMICS_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'infodynamics.jar')

    @classmethod
    def start(self) -> None:
        """
        Start Java Virtual Machine to run Java code inside this Python repository
        """
        if not jp.isJVMStarted():
            logging.info('Starting JVM...')
            try:
                jp.startJVM(jp.getDefaultJVMPath(), '-ea',
                            f"-Djava.class.path={self.INFODYNAMICS_PATH}")
                logging.info('Done.')
            except:
                logging.fatal('Error starting JVM')
                sys.exit(0)

    @classmethod
    def stop(self) -> None:
        """
        Gracefully exit Java Virtual Machine
        """
        if jp.isJVMStarted():
            logging.info('Shutting down JVM...')
            jp.shutdownJVM()
            logging.info('Done.')

    @classmethod
    def check(self) -> None:
        """
        Check if Java Virtual Machine is running.
        """
        if not jp.isJVMStarted():
            logging.error('JVM not started. Please start explicitly.')
            sys.exit(0)


    @classmethod
    def javify(self, X: np.ndarray, dtype: jp.JClass) -> jp.JArray:
        """
        Convert a numpy array into a Java array to pass to the JIDT classes and
        functions.
        Given a 1-dim np array of shape (D,), return Java array[] of size D
        Given a 2-dim np array of shape (1, D), return Java array[] of size D
        Given a 2-dim np array of shape (D1, D2), return Java array[][] of size D1 x D2

        Params
        ------
        X
            numpy array of shape (D,) or (D1,D2) representing a time series

        Returns
        ------
        jX
            the X array cast to Java Array
        """
        X = np.array(X)

        if len(X.shape) > 1 and X.shape[1] == 1:
            X = X.reshape((X.shape[0],))

        if len(X.shape) == 1:
            dim = 1
            X = X[np.newaxis, :]
        else:
            dim = len(X.shape)

        if dim > 1:
            jX = jp.JArray(dtype, dim)(X.tolist())
        else:
            # special case to deal with scalars
            jX = jp.JArray(dtype, 1)(X.flatten())

        return jX

