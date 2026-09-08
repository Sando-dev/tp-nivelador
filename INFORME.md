# Informe TP Nivelador

## Protocolo de Comunicación

Para la comunicación entre los clientes y el servidor se implementó un protocolo propio sobre TCP. Cada mensaje comienza con un encabezado que indica el tipo de mensaje y el tamaño del payload. Para esto se usa 1 byte para el tipo de mensaje y 2 bytes en Big Endian para la longitud del payload.

Para serializar cada apuesta se utilizó un esquema TLV (Type-Length-Value). De esta forma, cada campo se envía indicando primero qué tipo de campo es, luego su longitud y finalmente su valor. Esto permite reconstruir del otro lado datos como la agencia, nombre, apellido, documento, fecha de nacimiento y número apostado.

Las apuestas no se envían de a una, sino agrupadas en batches para reducir la cantidad de mensajes. Cada batch contiene varias apuestas y cada una está precedida por su longitud, lo que permite separarlas correctamente al recibirlas. La cantidad de apuestas por batch se configura mediante BATCH_SIZE.

Cuando el servidor procesa correctamente todas las apuestas de un batch responde con BATCH_OK. Si ocurre algún error responde con BATCH_ERROR. Cuando el cliente termina de enviar todas sus apuestas manda un mensaje FINISH. Luego el servidor devuelve los ganadores mediante mensajes WINNER y finalmente envía otro FINISH para indicar que terminó la respuesta.


## Concurrencia y sincronización

El servidor fue implementado usando multithreading. Cada vez que se conecta un cliente se crea un thread independiente para atenderlo, por lo que varias agencias pueden enviar y procesar apuestas al mismo tiempo.

Como hay información compartida entre distintos threads, se utilizan mecanismos de sincronización. El acceso al almacenamiento de apuestas se protege con un Lock, evitando que dos threads lean o escriban al mismo tiempo de forma inconsistente.

También se utiliza una Condition para manejar el quorum de agencias. Cuando una agencia termina de enviar sus apuestas, se agrega su identificador al conjunto de agencias finalizadas. Los threads esperan hasta que la cantidad de agencias terminadas alcanza el valor de AGENCY_QUORUM_MIN. Cada vez que una nueva agencia termina, se notifica a los threads que están esperando para que vuelvan a comprobar si ya se alcanzó el quorum.

Una vez alcanzado el quorum, cada cliente recibe solamente los ganadores que corresponden a su propia agencia.