# How to generate XML structs

	git clone --depth=1 https://gitlab.eclipse.org/eclipse/titan/titan.core.git $HOME/titan.core
	go install github.com/xuri/xgen/cmd/...@latest
	xgen -p titan -i $HOME/titan.core/etc/xsd/TPD.xsd -o titan_gen -l Go


