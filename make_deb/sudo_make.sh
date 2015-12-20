#! /bin/bash -e
cd make_deb
chown root dcoin/usr/share/dcoin/dcoin
chgrp root dcoin/usr/share/dcoin/dcoin
chown root dcoin64/usr/share/dcoin/dcoin
chgrp root dcoin64/usr/share/dcoin/dcoin
dpkg-deb --build dcoin
dpkg-deb --build dcoin64
zip -j dcoin_linux32.zip dcoin/usr/share/dcoin/dcoin
zip -j dcoin_linux64.zip dcoin64/usr/share/dcoin/dcoin
mv dcoin_linux32.zip /home/z/multiplatform/dc-compiled/dcoin_linux32.zip
mv dcoin_linux64.zip /home/z/multiplatform/dc-compiled/dcoin_linux64.zip
mv dcoin.deb /home/z/multiplatform/dc-compiled/dcoin_linux32.deb
mv dcoin64.deb /home/z/multiplatform/dc-compiled/dcoin_linux64.deb
rm -rf dcoin64/usr/share/dcoin/dcoin
rm -rf dcoin/usr/share/dcoin/dcoin
