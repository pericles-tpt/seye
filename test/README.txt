# IMPORTANT NOTE #
For testing we are currently using hardcoded values for what we expect to be
in `testDir/`.

Please do NOT modify the original files in `testDir/` i.e:
testDir/
    A/
        A
        B
        C
        D
    B/
    C/
    D/
    A1
    B1
    C1

If you wish to test MODIFYING `testDir/`, create some temporary files for your
tests and remove them after your test is complete (you can use `defer` to 
ensure that happens, e.g. `defer os.RemoveAll("./testDir/A/MyTempDir")`)