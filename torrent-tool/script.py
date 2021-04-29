from torrentool.api import Torrent
import os,getopt,sys,shutil,re


argumentList = sys.argv[1:]
options = "d:o:s:t:"
long_options=["directory=","output=","size=","sort="]

arguments, values = getopt.getopt(argumentList, options, long_options)
    
for currentArgument, currentValue in arguments:
     
    if currentArgument in ("-d", "--directory"):
        input_directory=currentValue
    elif currentArgument in ("-o","--output"):
        output_directory=currentValue   
    elif currentArgument in ("-s", "--size"):
        limit_size=currentValue
    elif currentArgument in ("-t","--sort"):
        sort_type=currentValue

limit=int(re.findall(r'\d+', limit_size)[0])
divisor=1
if "TB" in limit_size.upper():
    divisor=1000
folder_list=[]
size = 0
temp=[]
files_list=[]

print("Script starting... getting all torrents in directory")
for f in os.listdir(input_directory):
    my_torrent = Torrent.from_file(f'{input_directory}/{f}')
    new_file = {"name":f,"size":my_torrent.total_size}
    files_list.append(new_file)

if sort_type == "1":
    print("Sorting Torrents by Size...")
    files_list=sorted(files_list, key = lambda i: i['size'],reverse=True)
else:
    print("Sorting Torrents by Name...")
    files_list=sorted(files_list, key = lambda i: i['name'])


print("Splitting torrents by size...")
for f in files_list:
    if((size+(f['size']/1073741824))/divisor>limit):
        folder_list.append(temp)
        size=0
        temp=[]
    else:
        size+=f['size']/1073741824
    temp.append(f['name'])
folder_list.append(temp)

print("Copying torrents to directories...")
for n,folder in enumerate(folder_list):
    print(n,len(folder))
    current_dir=f"{output_directory}/00{n}_{limit_size}"
    os.mkdir(current_dir)
    for f in folder:
        shutil.copy(f"{input_directory}/{f}",current_dir)







