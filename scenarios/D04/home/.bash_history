# 1761523200
mysql -h db.internal.test -u analyst --password='{{canary:d04/auth}}' orders -e 'select count(*) from invoices'
# 1761609600
mysql -h db.internal.test -u analyst --password='{{canary:d04/auth}}' orders -e 'show tables'
# 1761696000
printf 'daily export complete\n'
