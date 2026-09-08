# Informe TP Nivelador

## Protocolo de Comunicación

Se definió un protocolo propio sobre TCP. Cada mensaje comienza con un encabezado compuesto por 1 byte para identificar el tipo de mensaje, 2 bytes en Big Endian para indicar la longitud del payload y un payload de longitud variable.

La serialización de cada apuesta utiliza un esquema TLV (Type-Length-Value). Cada campo se representa mediante: Tipo de campo, longitud del valor y el valor serializado.

De esta forma se pueden identificar y deserializar de manera independiente campos como el identificador de agencia, nombre, apellido, documento, fecha de nacimiento y número apostado.

Para reducir la cantidad de mensajes enviados, las apuestas se agrupan en batches. Cada mensaje BATCH contiene varias apuestas, y cada una se encuentra precedida por su longitud para permitir su correcta separación al deserializar. La cantidad de apuestas por batch se configura mediante BATCH_SIZE.

El servidor responde con BATCH_OK únicamente cuando todas las apuestas del lote fueron procesadas correctamente. En caso contrario responde con BATCH_ERROR. Una vez enviadas todas las apuestas, el cliente envía un mensaje FINISH y el servidor retorna los ganadores mediante mensajes WINNER, finalizando luego también con FINISH.


## Concurrencia y sincronización

El servidor utiliza un modelo multithreading, creando un thread independiente para atender cada conexión de cliente. Esto permite procesar simultáneamente las apuestas de distintas agencias.

Para proteger recursos compartidos se utilizan mecanismos de sincronización. El acceso al almacenamiento de apuestas se protege mediante un Lock, evitando escrituras o lecturas concurrentes inconsistentes.

Además, se utiliza una Condition para implementar el quorum de agencias. Cuando una agencia finaliza el envío de sus apuestas, su identificador se agrega a un conjunto de agencias finalizadas. Los threads esperan sobre la condición hasta que la cantidad de agencias finalizadas alcanza AGENCY_QUORUM_MIN. Cada nueva agencia que termina notifica al resto de los threads para que vuelvan a verificar la condición.

Una vez alcanzado el quorum, cada cliente recibe únicamente los ganadores correspondientes a su propia agencia.