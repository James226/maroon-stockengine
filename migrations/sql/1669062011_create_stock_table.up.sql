CREATE TABLE [Stock] (
    [Item] varchar(20),
    [Location] varchar(20),
    [LocationStock] decimal,
    [ReserveStock] decimal,
    [FinalStock] decimal,
    [UOM] varchar(5),
    CONSTRAINT PK_Stock PRIMARY KEY ([Item], [Location])
)