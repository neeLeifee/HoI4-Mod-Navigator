"""Generate labelled product illustrations (not Windows screenshots). Requires Pillow."""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import math
OUT=Path(__file__).resolve().parents[1]/'assets'
OUT.mkdir(exist_ok=True)
F='/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf'; B='/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf'
NAVY='#182839'; INK='#25364a'; MUTED='#64748b'; GOLD='#e0ad60'; BLUE='#356c9e'; WHITE='#ffffff'; BG='#edf1f5'
def font(n,b=False):return ImageFont.truetype(B if b else F,n)
def text(d,x,y,s,n=20,b=False,c=INK):d.text((x,y),s,font=font(n,b),fill=c,anchor='lt')
def box(d,x,y,w,h,fill=WHITE,outline=None,r=12):d.rounded_rectangle((x,y,x+w,y+h),r,fill=fill,outline=outline,width=1)
def folder(d,x,y,scale=1):
 d.rounded_rectangle((x,y,x+26*scale,y+18*scale),radius=2,fill='#d69a45');d.rectangle((x,y,x+11*scale,y+5*scale),fill='#d69a45');d.rounded_rectangle((x,y+4*scale,x+26*scale,y+20*scale),radius=2,fill='#efc477')
def file(d,x,y):
 d.rounded_rectangle((x,y,x+16,y+21),2,fill='#ffffff',outline='#8a9bac');d.line((x+4,y+8,x+12,y+8),fill='#8a9bac');d.line((x+4,y+12,x+12,y+12),fill='#8a9bac')
def canvas(w,h):return Image.new('RGB',(w,h),BG)
def mark(d,w,h):text(d,30,h-28,'ILLUSTRATION • Simulated UI, not a Windows recording',14,False,MUTED)
def brand():
 im=Image.new('RGB',(1280,640),NAVY);d=ImageDraw.Draw(im)
 for i in range(6):d.line((820+i*70,0,570+i*70,640),fill='#23394b',width=1)
 box(d,76,68,94,94,'#2c4054',r=18);folder(d,96,100,2);d.line((113,116,133,116,133,132,151,132),fill=NAVY,width=4);d.polygon([(145,125),(156,132),(145,139)],fill=NAVY)
 text(d,76,203,'HoI4 Mod Navigator',54,True,WHITE);text(d,80,284,'Vanilla ↔ mods',32,True,GOLD);text(d,80,340,'Jump to matching paths. Keep modding.',24,False,'#d6e1ec')
 for i,(s,sub) in enumerate([('01','Context-aware menus'),('02','Two-way navigation'),('03','Copy & compare')]):
  x=80+i*382;box(d,x,460,350,99,'#24384c',r=12);text(d,x+20,478,s,16,True,GOLD);text(d,x+20,505,sub,21,True,WHITE)
 im.save(OUT/'social-preview.png');im.resize((1120,560),Image.Resampling.LANCZOS).save(OUT/'banner.png')
def explorer(d,x,y,w,h,title,path,rows,selected=None):
 box(d,x,y,w,h,WHITE,'#b9c7d5',9);d.rectangle((x+1,y+30,x+w-1,y+31),fill='#d7dee7');folder(d,x+13,y+7,.8);text(d,x+42,y+7,title,16,True);text(d,x+w-68,y+7,'–  □  ×',15,c=MUTED)
 box(d,x+12,y+44,w-24,35,'#f5f7fa','#d5dee7',5);text(d,x+22,y+53,path,15)
 d.rectangle((x+1,y+91,x+132,y+h-1),fill='#f3f5f8');text(d,x+15,y+105,'Quick access',14,True);text(d,x+15,y+137,'Desktop',14,c=MUTED);text(d,x+15,y+166,'Documents',14,c=MUTED);text(d,x+15,y+195,'This PC',14,c=MUTED)
 text(d,x+149,y+103,'Name',14,c=MUTED);d.line((x+139,y+127,x+w-10,y+127),fill='#e4e9ef')
 for i,(s,isfolder) in enumerate(rows):
  yy=y+140+i*32
  if i==selected:box(d,x+141,yy-4,w-151,30,'#d9eafb',r=3)
  (folder if isfolder else file)(d,x+152,yy);text(d,x+184,yy+1,s,16)
def pointer(d,x,y):d.polygon([(x,y),(x,y+26),(x+7,y+18),(x+13,y+30),(x+19,y+27),(x+12,y+15),(x+24,y+15)],fill=WHITE,outline='#16283b')
def demo(step):
 im=canvas(1200,700);d=ImageDraw.Draw(im);text(d,30,25,'From vanilla to your mod',30,True);text(d,30,65,'Same relative path • separate Explorer window • missing folders created',18,c=MUTED)
 explorer(d,30,110,550,390,'Hearts of Iron IV','Hearts of Iron IV  ›  gfx  ›  interface',[('technologies',True),('ideas',True),('equipmentdesigner',True),('interface.txt',False)])
 if step>=2:
  explorer(d,620,110,550,390,'FFU2','FFU2  ›  gfx  ›  interface',[('technologies',True),('ideas',True)])
 else:
  box(d,620,110,550,390,'#e1e7ee','#c8d3df');text(d,654,270,'FFU2 destination',23,True,MUTED);text(d,654,306,'Opens after you choose the command.',17,c=MUTED)
 if step==1:
  box(d,225,250,340,154,'#f4f5f6','#a7b4c0',3)
  text(d,242,262,'Вид    ›',15);text(d,242,290,'Обновить',15);box(d,230,316,330,31,'#d9eafb',r=2);text(d,239,324,'Открыть идентичную папку в «FFU2»',14);text(d,242,365,'Свойства',15);pointer(d,530,332)
 elif step==0:pointer(d,430,290)
 else:
  d.line((581,300,618,300),fill=BLUE,width=4);d.polygon([(618,300),(607,293),(607,307)],fill=BLUE)
 descriptions=[('1','Right-click in the vanilla folder'),('2','Choose the named mod command'),('3','The matching mod folder opens')]
 for i,(n,s) in enumerate(descriptions):
  x=30+i*385;box(d,x,536,370,96,WHITE if i!=step else '#d9eafb',r=10);text(d,x+17,553,n,18,True,BLUE);text(d,x+17,584,s,17,True)
 mark(d,1200,700);return im

def reverse():
 im=canvas(1200,650);d=ImageDraw.Draw(im);text(d,30,25,'Find the original file',30,True);text(d,30,65,'Right-click a mod file → reveal the corresponding vanilla file',18,c=MUTED)
 explorer(d,30,105,550,400,'FFU2','FFU2  ›  common  ›  units',[('infantry.txt',False),('artillery.txt',False),('support.txt',False)],0)
 explorer(d,620,105,550,400,'Hearts of Iron IV','Hearts of Iron IV  ›  common  ›  units',[('artillery.txt',False),('infantry.txt',False),('support.txt',False)],1)
 box(d,177,270,400,104,'#f4f5f6','#a7b4c0',3);box(d,181,278,392,32,'#d9eafb',r=2);text(d,190,285,'Показать файл в ванильной игре (из «FFU2»)',14);text(d,190,327,'Сравнить с «FFU2» / оригиналом (VS Code)',14);pointer(d,539,294)
 d.line((581,300,618,300),fill=BLUE,width=4);d.polygon([(618,300),(607,293),(607,307)],fill=BLUE)
 box(d,30,531,1140,72,WHITE,r=10);text(d,50,550,'Original file selected in a new window; no folders or files are created in vanilla.',20,True);mark(d,1200,650);im.save(OUT/'reverse-navigation.png')
def settings():
 im=canvas(1200,830);d=ImageDraw.Draw(im);text(d,30,22,'Configure once. Navigate from Explorer.',30,True);text(d,30,61,'Russian-language settings UI • example paths',18,c=MUTED)
 x,y=36,104;box(d,x,y,790,682,WHITE,'#bdcbd9',10);text(d,x+18,y+13,'HoI4 Mod Navigator  •  v2.0',16,True);d.line((x,y+42,x+790,y+42),fill='#dce3eb');text(d,x+24,y+55,'HoI4 Mod Navigator',26,True);text(d,x+24,y+92,'v2.0 / Настройки интеграции',14,c=MUTED)
 def field(xx,yy,w,s):box(d,x+xx,y+yy,w,30,'#f7f9fb','#c5d0dc',3);text(d,x+xx+8,y+yy+7,s,13)
 def btn(xx,yy,w,s,primary=False):box(d,x+xx,y+yy,w,30,BLUE if primary else '#edf2f7','#c5d0dc',4);text(d,x+xx+8,y+yy+7,s,13,c=WHITE if primary else INK)
 text(d,x+24,y+130,'01   Оригинальная игра',18,True);field(24,161,594,r'D:\Steam\steamapps\common\Hearts of Iron IV');btn(630,161,132,'Выбрать…')
 text(d,x+24,y+210,'02   Мои моды',18,True);box(d,x+24,y+240,738,72,'#f7f9fb','#c5d0dc',3);text(d,x+36,y+252,r'FFU2    —    D:\Mods\FFU2',15);text(d,x+36,y+280,r'Kaiserreich    —    D:\Mods\Kaiserreich',15)
 text(d,x+24,y+323,'Название',13);text(d,x+250,y+323,'Корневая папка мода',13);field(24,346,214,'FFU2');field(250,346,368,r'D:\Mods\FFU2');btn(630,346,132,'Выбрать…');btn(24,387,143,'Добавить мод');btn(178,387,160,'Сохранить правки');btn(349,387,168,'Убрать из списка')
 text(d,x+24,y+440,'03   Команды Проводника',18,True)
 for i,s in enumerate(['Открыть папку в моде','Скопировать файл в мод','Показать папку / файл в ванилле','Сравнить файлы (VS Code)','Создавать недостающие папки в моде','Фильтровать команды ПКМ','Совместимый режим (старое меню без DLL)']):
  xx=x+24+(i%2)*370; yy=y+471+(i//2)*27;box(d,xx,yy,16,16,BLUE if i<6 else WHITE,'#a4b4c5',2)
  if i<6:d.line((xx+3,yy+8,xx+7,yy+12,xx+13,yy+4),fill=WHITE,width=2)
  text(d,xx+23,yy+1,s,13)
 btn(24,603,220,'Применить настройки',True);btn(260,603,204,'Удалить интеграцию');btn(630,603,132,'Закрыть')
 for yy,num,head,sub in [(190,'01','Game root','Contains hoi4.exe'),(340,'02','Mod roots','Up to 20 mods'),(535,'03','Integration','Apply or roll back')]:
  box(d,850,yy,320,105,'#ffffff',r=9);text(d,868,yy+13,num,16,True,BLUE);text(d,868,yy+39,head,21,True);text(d,868,yy+69,sub,16,c=MUTED)
 mark(d,1200,830);im.save(OUT/'settings-illustration.png')
brand();settings();reverse();frames=[demo(i) for i in range(3)];frames[2].save(OUT/'folder-navigation.png');frames[0].save(OUT/'navigation-demo.gif',save_all=True,append_images=frames[1:],duration=[1600,2200,2600],loop=0,optimize=True)
