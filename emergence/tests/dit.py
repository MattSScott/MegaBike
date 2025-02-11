import dit
from dit.shannon import mutual_information as mi
import numpy as np

from emergence.dit.pid import pid_psi
from emergence.calc.dit import DitDistCalc

if __name__ == "__main__":
    stats = []
    for n_var in range(3, 6):
        for q in range(0, n_var):
            diffs = []
            for _ in range(20):
                d = dit.random_distribution(n_var, 2)
                i = dit.pid.PID_MMI(d)
                ppid = pid_psi(i, q = q)
                calc = DitDistCalc(d, mi)
                pmi  = calc.psi(q = q)
                diffs.append(np.abs(ppid - pmi))
            m = np.mean(diffs)
            s = np.std(diffs)
            print(f"n={n_var}\tq={q}\t{m}\t{s}")

