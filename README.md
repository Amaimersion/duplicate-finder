# duplicate-finder

This program is intended to find same files across two folders. Files are being compared by MD5 hash. Duplicates will be handled: printed in console, moved in another folder, etc.

## How to execute

Download and execute binary file using terminal. Use `--help` flag to see help message.

## Notes

- Before using `--move` flag, first execute the program without that flag. You will see files that will be touched by the program. This way you will be able to check if the program behaves as you expected.
- `--f1` and `--f2` can point to the same folder.
- This program will not delete or rename any files to avoid incidents.
- `--move` will keep original folder structure.
- RAM usage will be efficient even for large amount of files.

## Examples

```
./duplicate-finder --f1 ~/f1 --f2 ~/f2
```

This will compare two folders and log information in console.

```
duplicate-finder.exe --f1 "D:\user\Desktop\folder 1" --f2 "D:\user\Desktop\folder 2" --move "D:\user\Desktop\Duplicates"
```

This will compare two folders and move all duplicates in separate folder.

```
./duplicate-finder -f1 /home/user/f1 -f2 /home/user/f1 -move /home/user/dups -output /home/user/logs.txt
```

This will find same files in entire same folder. All duplicates will be moved in separate folder. Original folder will left only with original files. All logs will be written in separate file.
