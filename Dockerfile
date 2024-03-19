FROM kalilinux/kali-rolling:latest
RUN apt-get update && \
    apt-get install -y git python3-pip figlet sudo && \
    apt-get install -y boxes php curl xdotool wget

WORKDIR /root/introlabs
COPY requirements.txt ./
RUN pip3 install --no-cache-dir boxes flask lolcat requests -r requirements.txt
COPY . .
RUN true && echo "/root/introlabs/" > /home/introlabspath.txt;
EXPOSE 1-65535
ENTRYPOINT ["python3", "/root/introlabs/introlabs.py"]