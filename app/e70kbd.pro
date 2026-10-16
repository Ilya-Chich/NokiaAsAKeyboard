# ProGuard preverify only. The E70 VM (S60 3rd, not FP1) does not start a
# suite whose screen class was renamed into the default package.

-microedition
-dontoptimize
-dontshrink
-dontobfuscate

-keep public class * extends javax.microedition.midlet.MIDlet
