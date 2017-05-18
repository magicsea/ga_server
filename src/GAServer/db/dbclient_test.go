package db

import (
	"testing"

	"fmt"

	"github.com/astaxie/beego/orm"
)

type User struct {
	Id      int
	Name    string
	Profile *Profile `orm:"rel(one)"`      // OneToOne relation
	Post    []*Post  `orm:"reverse(many)"` // 设置一对多的反向关系
}

type Profile struct {
	Id   int
	Age  int16
	User *User `orm:"reverse(one)"` // 设置一对一反向关系(可选)
}

type Post struct {
	Id    int
	Title string
	User  *User  `orm:"rel(fk)"` //设置一对多关系
	Tags  []*Tag `orm:"rel(m2m)"`
}

type Tag struct {
	Id    int
	Name  string
	Posts []*Post `orm:"reverse(many)"`
}

//主键`orm:"pk"`
//唯一`orm:"unique"`
//忽略`orm:"-"`
//允许空`orm:"null"`
//varchar大小`orm:"size(60)"
//默认`orm:"default(1)"`
type Player struct {
	Id      int    `orm:"column(uid)",auto`
	Name    string `orm:"column(username)"`
	Cgid    int
	Lv      int
	Exp     int
	Exptime int64
}

func TestDB_rui_normal(t *testing.T) {
	orm.RegisterModel(new(Player))
	//连接数据库
	client, err := ConnectDB("root:tcg123456@tcp(192.168.3.194:3306)/test", "default")
	if err != nil {
		t.Error(err)
		return
	}
	//增加记录
	/*
		sql := fmt.Sprintf("insert into player values(1, ?, 2, 3, 4, 342423423)")
		_, err1 := client.Raw(sql, []string{"a"}).Exec()
		if err1 != nil {
			t.Error(err1)
			return
		}
		for i := 2; i <= 10; i++ {
			sql = fmt.Sprintf("insert into player values(?, \"ufo\", 2, ?, 4, 342423423)")
			_, err1 = client.Raw(sql, []int{i, i + 1}).Exec()
			if err1 != nil {
				t.Error(err1)
				return
			}
		}
	*/
	//查询
	/*
		p := new(Player)
		sql1 := fmt.Sprintf("select * from player where uid = ?")
		err = client.Raw(sql1, []int{1}).QueryRow(&p)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("select uid:%d, name :%s, Cgid :%d, \nlv :%d, exp :%d, exptime:%u", p.Id, p.Name, p.Cgid, p.Lv, p.Exp, p.Exptime)
		fmt.Printf("\nplayer uid is 1 %v\n", p)
		ps := make([]Player, 6)
		//var Player []ps
		sql1 = fmt.Sprintf("select * from player where uid >= ?")
		num, errr := client.Raw(sql1, []int{5}).QueryRows(&ps)
		if errr != nil {
			t.Error(errr)
			return
		} else {
			fmt.Printf("共查询到 %d 个玩家\n", num)
			for index, _ := range ps {
				fmt.Printf("\n第 %d 玩家信息为：%v", index+1, ps[index])
			}
			fmt.Printf("\n")
		}
	*/
	//删除某些记录
	/*
		sql = fmt.Sprintf("delete from player WHERE uid = ?;")
		fmt.Println(sql)
		_, err3 := client.Raw(sql, []int{1}).Exec()
		if err3 != nil {
			t.Error(err3)
			return
		}
	*/
	//修改记录
	sqlUpdate := fmt.Sprintf("update player set lv = ?")
	_, err = client.Raw(sqlUpdate, []int{111}).Exec()
	if err != nil {
		t.Error(err)
		return
	}

}

func TestDB_ruiy(t *testing.T) {
	//数据库 表注册
	orm.RegisterModel(new(Player))
	//连接数据库
	client, err := ConnectDB("root:tcg123456@tcp(192.168.3.194:3306)/test", "default")
	if err != nil {
		t.Error(err)
		return
	}
	p0 := new(Player)
	client.Delete(p0)
	//增加记录
	/*
		var p1 *Player = &Player{Id: 1, Name: "a", Cgid: 1, Lv: 2, Exp: 12, Exptime: time.Now().Unix()}
		var p2 *Player = &Player{Id: 2, Name: "b", Cgid: 2, Lv: 3, Exp: 13, Exptime: time.Now().Unix()}
		//插入数据
		fmt.Println(client.Insert(p1))
		fmt.Println(client.Insert(p2))
		players := []Player{
			{Id: 3, Name: "c", Cgid: 3, Lv: 4, Exp: 14, Exptime: time.Now().Unix()},
			{Id: 4, Name: "d", Cgid: 4, Lv: 5, Exp: 15, Exptime: time.Now().Unix()},
			{Id: 5, Name: "e", Cgid: 5, Lv: 6, Exp: 16, Exptime: time.Now().Unix()},
		}
		//批量插入
		client.InsertMulti(3, players)
	*/
	//删除记录
	/*
		p3 := new(Player)
		p3.Id = 4
		client.Delete(p3)
	*/

	//修改记录
	/*
		player1 := new(Player)
		player1.Id = 2
		player1.Name = "change name"
		client.Update(player1, "Name")
		err2 := client.Read(player1)
		if err2 != nil {
			t.Error(err2)
			return
		}
		fmt.Println("\n update end select is :", player1)
	*/
	//查询记录
	/*
		player := new(Player)
		player.Id = 1
		err1 := client.Read(player)
		if err1 != nil {
			t.Error(err1)
			return
		}
		fmt.Println("\nselect data is :", player)
	*/
}

func TestDB(t *testing.T) {
	orm.RegisterModel(new(Player))
	client, err := ConnectDB("root:tcg123456@tcp(192.168.3.194:3306)/test", "default")
	if err != nil {
		t.Error(err)
		return
	}
	p := Player{Id: 6, Name: "www", Exp: 1}
	client.Insert(&p)
	if err := client.Read(&p, "Exp"); err != nil {
		t.Error(err)
		return
	}
	p.Name = "yyyy"
	client.Update(&p, "Name")
	//client.Delete(&p, "Exp")
	t.Log("insert ok :", p)
}

/*
func main() {
	o := orm.NewOrm()
	o.Using("default") // 默认使用 default，你可以指定为其他数据库

	profile := new(Profile)
	profile.Age = 30

	user := new(User)
	user.Profile = profile
	user.Name = "slene"

	fmt.Println(o.Insert(profile))
	fmt.Println(o.Insert(user))

}
*/
