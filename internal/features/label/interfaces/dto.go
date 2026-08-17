package interfaces

type OrderRequest struct {
	CollectionID uint   `json:"collection_id"`
	Email        string `json:"email" validate:"omitempty,email,max=200"`
	Format       string `json:"format" validate:"omitempty,oneof=5160 8160 5162 5163 6871"`
}
