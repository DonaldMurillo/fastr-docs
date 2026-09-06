---
tags: [ejemplos, pantallas, superficies]
locale: es
translation_of: /examples/custom-surface
---

# Lista para una superficie propia

Una superficie de documentación a medida sigue siendo una ruta. Empieza por la
tarea de la persona, elige la estrategia de renderizado y mantén la ruta
conectada al modelo de navegación y búsqueda que comparte el proyecto.

## Decide el contrato de la ruta

1. Nombra la tarea en el título y la descripción de la ruta.
2. Elige una página de Markdown para explicar algo duradero, o una pantalla
   tipada para interactuar.
3. Asigna un orden explícito entre hermanas.
4. Decide si la superficie funciona sin conexión.
5. Añade etiquetas y texto de búsqueda cuando el texto visible no baste.

## Componer con GoFastr

Usa los controles del framework para formularios, pestañas, desplegables, vistas
de datos, notificaciones y comportamiento adaptable. Deja el estado y las
acciones de servidor en el componente de pantalla. Que la envoltura de
documentación siga gobernando la cabecera, el cajón, la paleta de comandos, la
navegación interna y el tema.

## Verifica el recorrido

Prueba la ruta en anchos de escritorio y de móvil. Cubre el primer renderizado,
la interacción principal, los estados de error o vacíos, salir y volver, el
acceso por teclado y la exportación estática. El
[playground](/examples/playground) muestra este patrón en una pantalla pequeña.
