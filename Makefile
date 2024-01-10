all:
	go build
	rm -rf output.txt output1.txt output2.txt tmp
	mkdir -p tmp
	./foli2
	[ -f output.txt ] && (cat output.txt | jq > a && mv -f a output.txt) || true
	[ -f output1.txt ] && (cat output1.txt | jq > a && mv -f a output1.txt) || true
	[ -f output2.txt ] && (cat output2.txt | jq > a && mv -f a output2.txt) || true
