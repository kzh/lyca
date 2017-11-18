# lyca
Lyca is a programming language that I am currently working on. The main goal for this project is to help me learn compiler design. Being in a very early development stage, Lyca is full of bugs, but here is a demo of some of the functioning features in the form of a Linked List implementation in Lyca.

```
func () > main > () {
    List list = make List < ();
    
    list.append(make Node < (10));
    list.append(make Node < (4));
    list.append(make Node < (3));

    list.print();
}

tmpl List {
    Node head;
    int length;

    constructor < () {
        this.length = 0;
    }

    func (Node node) > append > () {
        if (this.length == 0) {
            this.head = node;
        } else {
            Node last = this.get(this.length - 1);
            last.next = node;
        }

        this.length = this.length + 1;
    }

    func (int depth) > get > (Node) {
        Node node = this.head;
        for (; depth != 0; depth = depth - 1) {
            node = node.next;
        }

        return node;
    }

    func () > print > () {
        Node node = this.head;
        for (int i = 0; i != this.length; i = i + 1) {
            printf("Index: %d Value: %d \n", i, node.value);
            node = node.next;
        }
    }
}

tmpl Node {
    Node next;
    int value;

    constructor < (int val) {
        this.value = val;
    }
}
```


![alt tag](https://i.imgur.com/Vqqgm81.png)
