package responses

var enmapscode = map[int]string{
	0:     "Success",
	1:     "Failed",
	1000:  "Authorization header missing",
	1001:  "Invalid Token",
	1003:  "Bad Request",
	1004:  "Unauthorized User",
	1005:  "Invalid generate signature",
	1006:  "Failed create session",
	1007:  "No Data Found",
	1008:  "Failed inserting data",
	1009:  "User cannot deleting, because has been order",
	1010:  "Cannot deleting this order",
	1011:  "Cannot updating this order",
	1012:  "Cannot pay this order",
	-1018: "Order not found",
}

func GetErrorCodeEN(code int) string {
	return enmapscode[code]
}

var idmapscode = map[int]string{
	0:     "Berhasil",
	1:     "Gagal",
	1000:  "Otorisasi tidak ada",
	1001:  "Token tidak sesuai",
	1003:  "Server tidak dapat mengenali permintaan Anda",
	1004:  "Izin pengguna tidak sah",
	1005:  "Gagal membuat tanda tangan",
	1006:  "Gagal membuat sesi",
	1007:  "Data tidak ditemukan",
	1008:  "Gagal menambahkan data",
	1009:  "User tidak bisa di hapus, karena user sudah memiliki data order",
	1010:  "Order ini tidak bisa di hapus",
	1011:  "Order ini tidak bisa di ubah",
	1012:  "Order ini tidak bisa di bayar",
	-1018: "Pesanan tidak ditemukan",
}

func GetErrorCodeID(code int) string {
	return idmapscode[code]
}
