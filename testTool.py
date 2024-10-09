import sys

def readSnapshotFileForProcess(snapshotFile):
    snapshots = []
    current_snapshot = {}
    current_snapshot["messages"] = {}
    with open(snapshotFile, 'r') as f:
        lines = f.readlines()
        for line in lines:
            if line.startswith("Process ID"):
                process_id = line.split(":")[1].strip()
                current_snapshot["process_id"] = process_id
            elif line.startswith("Snapshot ID"):
                snapshot_id = line.split(":")[1].strip()
                current_snapshot["snapshot_id"] = snapshot_id
            elif line.startswith("lcl"):
                lcl = line.split(":")[1].strip()
                current_snapshot["lcl"] = lcl
            elif line.startswith("Status"):
                status = line.split(":")[1].strip()
                current_snapshot["status"] = status
            elif line.startswith("Waiting"):
                waiting = line.split(":")[1].strip()
                waiting = waiting.strip("[]") 
                waiting_array = [value == "true" for value in waiting.split()]
                current_snapshot["waiting"] = waiting_array
            elif line.startswith("  Process"):
                split = line.split(",")
                process_id = split[0].split()[1]
                messages = split[1:]
                current_snapshot["messages"][process_id] = messages
            elif line.startswith("-------End Snapshot-------"):
                snapshots.append(current_snapshot)
                current_snapshot = {}
                current_snapshot["messages"] = {}
    return snapshots

# no máximo um processo na SC.
def inv1(snapshotsProcessList):
    in_mx_snapshots_ids = set()
    for snapshotList in snapshotsProcessList:
        for snapshot in snapshotList:
            if snapshot["status"] == "inMX":
                if snapshot["snapshot_id"] in in_mx_snapshots_ids:
                    print("Error: Two snapshots with the same snapshot id are inMx: " + snapshot["snapshot_id"])
                    return
                in_mx_snapshots_ids.add(snapshot["snapshot_id"])
    print("Invariant 1 is satisfied")

#  se todos processos estão em "não quero a SC", então todos waitings tem que ser falsos e não deve haver mensagens
def inv2(snapshotsProcessList):
    snapshots_len = len(snapshotsProcessList[0])
    for i in range(0, snapshots_len):
        all_no_mx = True
        for snapshotList in snapshotsProcessList:
            if snapshotList[i]["status"] != "noMX":
                all_no_mx = False
                break
        if all_no_mx:
            for snapshotList in snapshotsProcessList:
                if snapshotList[i]["waiting"] != [False, False, False]:
                    print("Error: All processes are not in MX but waiting is not false for snapshot id: " + snapshotList[i]["snapshot_id"])
                    return
                for process_id in snapshotList[i]["messages"]:
                    if len(snapshotList[i]["messages"][process_id]) != 0:
                        print("Error: All processes are not in MX but there are messages for snapshot id: " + snapshotList[i]["snapshot_id"])
                        return
    print("Invariant 2 is satisfied")

# se um processo q está marcado como waiting em p, então p está na SC ou quer a SC
def inv3(snapshotsProcessList):
    for snapshotList in snapshotsProcessList:
        for snapshot in snapshotList:
            for value in snapshot["waiting"]:
                if value == True:
                    if snapshot["status"] == "noMX":
                        print("Error: Some processes are waiting but not in MX for snapshot id: " + snapshot["snapshot_id"])
                        return
    print("Invariant 3 is satisfied")
                

def main():
    process_num = int(sys.argv[1])
    snapshots = []
    for i in range(0, process_num):
        snapshotFile = "snapshot" + str(i) + ".txt"
        snapshots.append(readSnapshotFileForProcess(snapshotFile))
    inv1(snapshots)
    inv2(snapshots)
    inv3(snapshots)
    

if __name__ == '__main__':
    main()