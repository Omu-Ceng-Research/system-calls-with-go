// 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
//
// Bu sınıf sistem çağrısı yapmamaktadır; yalnızca bellekteki veriyi tanımlamaktadır.
// Sunucu ve istemcideki iki kopyanın aynı olması gerekmektedir: serialVersionUID farklı
// olursa readObject InvalidClassException vermektedir.

import java.io.Serializable;

public class Person implements Serializable {
    private static final long serialVersionUID = 1L;
    // static: nesnenin değil sınıfın alanıdır, serileştirilmez.
    static String country = "TR";
    private int age;
    private String name;
    // transient: serileştirilmez, karşı tarafa 0 olarak ulaşır.
    transient int height;

    public int getAge() {
        return age;
    }

    public void setAge(int age) {
        this.age = age;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public int getHeight() {
        return height;
    }

    public void setHeight(int height) {
        this.height = height;
    }
}