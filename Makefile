all:
	go build
	rm -rf output.txt output1.txt output2.txt tmp
	mkdir -p tmp
	./foli2
	cat output1.txt | jq > a && mv -f a output1.txt
	cat output2.txt | jq > a && mv -f a output2.txt

