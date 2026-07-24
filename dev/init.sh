#!/usr/bin/env bash
# Statping Status Page Server for EC2

cd /home/ubuntu || exit 1
# shellcheck source=/dev/null
source /home/ubuntu/.profile 2> /dev/null || true
sudo rm -rf startup.sh > /dev/null
sudo curl -o startup.sh -H 'Cache-Control: no-cache' https://raw.githubusercontent.com/statping-ng/statping-ng/master/dev/startup.sh > /dev/null
sudo chmod +x startup.sh > /dev/null
sudo rm -f docker-compose.yml > /dev/null

EC2_ENDPOINT=$(curl -s http://169.254.169.254/latest/meta-data/public-hostname)
EC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)

if [ "$LETSENCRYPT_HOST" = "" ]; then
  sudo curl -o docker-compose.yml -H 'Cache-Control: no-cache' https://raw.githubusercontent.com/statping-ng/statping-ng/master/dev/docker-compose-single.yml > /dev/null
else
  printf "                    \n\n\n\nDomain found for SSL certificate - %s\n" "$LETSENCRYPT_HOST"
  printf "================================================================================================================\n"
  printf "You must set the domains DNS records to point to this server!\n"
  printf "   EC2 Server IP:     %s\n" "$EC_IP"
  printf "   EC2 Public DNS:    %s\n" "$EC2_ENDPOINT"
  printf "================================================================================================================\n"
  printf " CNAME    %s   =>   %s       (if not using Elastic IP)\n" "$LETSENCRYPT_HOST" "$EC2_ENDPOINT"
  printf "   A      %s   =>   %s                                        (or use A record if you are using an Elastic IP)\n" "$LETSENCRYPT_HOST" "$EC_IP"
  printf "================================================================================================================\n\n\n"

  sudo curl -o docker-compose.yml -H 'Cache-Control: no-cache' https://raw.githubusercontent.com/statping-ng/statping-ng/dev/docker-compose-ssl.yml > /dev/null
fi

sudo service docker start > /dev/null

if [ "$LETSENCRYPT_HOST" = "" ]; then
  sudo docker-compose pull > /dev/null
  sudo docker-compose up -d > /dev/null
else
  sudo LETSENCRYPT_HOST="$LETSENCRYPT_HOST" LETSENCRYPT_EMAIL="$LETSENCRYPT_EMAIL" docker-compose pull > /dev/null
  sudo LETSENCRYPT_HOST="$LETSENCRYPT_HOST" LETSENCRYPT_EMAIL="$LETSENCRYPT_EMAIL" docker-compose up -d > /dev/null
fi

sudo docker system prune -af > /dev/null

curl -s https://raw.githubusercontent.com/statping-ng/statping-ng/dev/init.sh | sudo tee /home/ubuntu/init.sh > /dev/null
sudo chmod +x /home/ubuntu/init.sh > /dev/null

printf "\n\n\n\n\n              Statping Status Page on EC2\n"
printf "================================================================================================================\n"
if [ "$LETSENCRYPT_HOST" = "" ]; then
  printf "Point your domain's DNS records to one of these endpoints.\n"
  printf "A RECORD     =>   %s   \n" "$EC_IP"
  printf "CNAME RECORD =>   %s\n" "$EC2_ENDPOINT"
  printf "================================================================================================================\n"
  printf "Your Statping Server is ready! Go to the URL below to begin.\n"
  printf "Statping URL: %s\n" "$EC2_ENDPOINT"
  printf "================================================================================================================\n"
else
  printf "Domain found for SSL certificate - %s\n" "$LETSENCRYPT_HOST"
  printf "================================================================================================================\n"
  printf "You must set the domains DNS records to point to this server!\n"
  printf "A RECORD     =>   %s   \n" "$EC_IP"
  printf "CNAME RECORD =>   %s\n" "$EC2_ENDPOINT"
  printf "================================================================================================================\n"
  printf "Once you set your DNS records, Lets Encrypt will automatically\n"
  printf "create a SSL certificate for you and redirect you to HTTPS\n\n"
  printf "================================================================================================================\n"
  printf "Your Statping Server is ready! Go to the URL below to begin.\n"
  printf "Statping URL: %s\n" "$EC2_ENDPOINT"
  printf "SSL Domain: %s\n" "$LETSENCRYPT_HOST"
  printf "================================================================================================================\n"
fi
