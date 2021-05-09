# APIs de Transporte Público en Santiago

## Saldo Bip!

Permite obtener el saldo de una tarjeta Bip!, consultándolo en el sitio de RedBip!.

Ejemplo: https://api.xor.cl/red/balance/87658765

## Itinerario de Paraderos

Permite obtener los buses que pasarán pronto en un paradero específico, consultando en el sitio oficial de Red.

Ejemplo: https://api.xor.cl/red/bus-stop/PA433

## Red de metro

Obtiene el estado de la red de metro según la página oficial. 

Los códigos de status de las estaciones significan lo siguiente:

* `0`: Estación operativa
* `1`: Estación Cerrada Temporalmente
* `2`: Estación no habilitada
* `3`: Accesos Cerrados


Ejemplo: https://api.xor.cl/red/metro-network